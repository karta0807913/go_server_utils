package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
	"gorm.io/gorm/schema"
)

var (
	typeName   = flag.String("type", "", "type name; must be set")
	method     = flag.String("method", "Find", "Create, Update, Find, First")
	apiPath    = flag.String("apiPath", "/", "API annotation for document")
	docDir     = flag.String("docDir", "../doc/", "Document prefix dir")
	required   = flag.String("require", "", "input required fields, default read gorm tag in struct which is not null without primaryKey and default")
	minItem    = flag.Uint("minItem", 1, "minimum number of items to choose")
	options    = flag.String("options", "", "input options fields")
	decoder    = flag.String("decoder", "json", "decoder: xml,json or etc")
	ignore     = flag.String("ignore", "", "which field should ignore")
	indexField = flag.String("indexField", "", "for an update index")
	max_limit  = flag.Uint("max_limit", 20, "search limit")
	min_limit  = flag.Uint("min_limit", 0, "min_limit")
)

func isDir(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

var namer = schema.NamingStrategy{}

var funcMap template.FuncMap = template.FuncMap{
	"tablename": func(name string) string {
		return namer.TableName(name)
	}, "to_pointer_type": func(field TemplateField) string {
		if field.StarExpr {
			return field.Type
		} else {
			return "*" + field.Type
		}
	}, "remove_pointer_type": func(field TemplateField) string {
		if field.StarExpr {
			return field.Type[1:]
		} else {
			return field.Type
		}
	}, "pluses": func(a, b int) interface{} {
		return a + b
	},
}

func main() {
	flag.Parse()
	if *typeName == "" {
		fmt.Fprintf(os.Stderr, "type required\n")
		flag.PrintDefaults()
		os.Exit(127)
	}
	if !NewCommaSet("Create,Update,Find,First").CheckAndDelete(*method) {
		log.Fatal("method must either Create, Update, Find or First, but got ", *method)
	}
	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}
	var rootDir string
	if isDir(args[0]) {
		rootDir = args[0]
	} else {
		rootDir = filepath.Dir(args[0])
	}

	requireSet := NewCommaSet(*required)
	optionsSet := NewCommaSet(*options)
	ignoreSet := NewCommaSet(*ignore)

	parsedPKG := parsePackage(*typeName)
	parsedTypes := parsedPKG.StructList
	if len(parsedTypes) == 0 {
		log.Fatal("can't find type ", *typeName)
	}

	//FINALLY, Generate Data
	filename := path.Join(rootDir,
		fmt.Sprintf(
			"%s_%s_route.go",
			strings.ToLower(*typeName),
			strings.ToLower(*method),
		))
	os.Remove(filename)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal("open file error ", err)
	}

	var parsedInformation interface{}
	var docFileName string
	if *method == "Create" {
		temp := MethodCreate(MethodCreateParams{
			ParsedType: parsedTypes[0],
			RequireSet: requireSet,
			OptionsSet: optionsSet,
			IgnoreSet:  ignoreSet,
			TagKey:     *decoder,
		})
		temp.Package = parsedPKG.Name
		t := template.New("").Funcs(funcMap)
		t = template.Must(t.Parse(CreateOrUpdateTemplate))
		t.Execute(file, temp)
		parsedInformation = temp
		docFileName = path.Join(*docDir, *method+temp.StructName+".go")
	} else if *method == "Update" {
		temp := MethodUpdate(MethodUpdateParams{
			ParsedType: parsedTypes[0],
			RequireSet: requireSet,
			OptionsSet: optionsSet,
			IgnoreSet:  ignoreSet,
			IndexField: *indexField,
			TagKey:     *decoder,
		})
		temp.Package = parsedPKG.Name
		t := template.New("").Funcs(funcMap)
		t = template.Must(t.Parse(CreateOrUpdateTemplate))
		t.Execute(file, temp)
		parsedInformation = temp
		docFileName = path.Join(*docDir, *method+temp.StructName+".go")
	} else if *method == "First" {
		temp := MethodSearch(MethodSearchParams{
			ParsedType: parsedTypes[0],
			RequireSet: requireSet,
			OptionsSet: optionsSet,
			IgnoreSet:  ignoreSet,
			TagKey:     *decoder,
		})
		temp.Package = parsedPKG.Name
		t := template.New("").Funcs(funcMap)
		t = template.Must(t.Parse(FirstTemplate))
		t.Execute(file, temp)
		parsedInformation = temp
		docFileName = path.Join(*docDir, *method+temp.StructName+".go")
	} else if *method == "Find" {
		temp := MethodSearch(MethodSearchParams{
			ParsedType: parsedTypes[0],
			RequireSet: requireSet,
			OptionsSet: optionsSet,
			IgnoreSet:  ignoreSet,
			TagKey:     *decoder,
		})
		temp.Package = parsedPKG.Name
		t := template.New("").Funcs(funcMap)
		t = template.Must(t.Parse(FindTemplate))
		t.Execute(file, temp)
		parsedInformation = temp
		docFileName = path.Join(*docDir, *method+temp.StructName+".go")
	} else {
		log.Fatal("method not support now :<")
	}
	file.Close()
	cmd := exec.Command("go", "fmt")

	if err := cmd.Run(); err != nil {
		fmt.Printf("go fmt code got error %s\n", err)
	}
	cmd = exec.Command("gopls", "imports", "-w", filename)
	if err := cmd.Run(); err != nil {
		fmt.Println("can't find gopls to import the code")
	}

	for key := range *requireSet {
		if key == "" {
			continue
		}
		fmt.Printf("warning: require field %s is not used\n", key)
	}
	for key := range *optionsSet {
		if key == "" {
			continue
		}
		fmt.Printf("warning: options field %s is not used\n", key)
	}

	// generate golang document files
	if _, err := os.Stat(*docDir); os.IsNotExist(err) {
		os.MkdirAll(*docDir, 0755)
	}

	file, err = os.OpenFile(docFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("open document file %s error, error is %s\n", docFileName, err)
	}
	err = GenerateDocument(*method, parsedInformation, file)
	if err != nil {
		fmt.Println(err)
	}
}

type Package struct {
	Name       string
	StructList []Type
}

type Type struct {
	ast.StructType
	Name   string
	Fields []Field
}

type Field struct {
	Name     string
	RawField ast.Expr
	Tag      reflect.StructTag
	DocList  []*ast.Comment
	Type     string
}

func getExprName(expr ast.Expr) string {
	var typeStr string
	switch t := expr.(type) {
	case *ast.Ident:
		typeStr = t.Name
	case *ast.SelectorExpr:
		typeStr = t.X.(*ast.Ident).Name + "." + t.Sel.Name
	case *ast.ArrayType:
		var prefix string
		if t.Len == nil {
			prefix = ""
		} else {
			prefix = t.Len.(*ast.BasicLit).Value
		}
		typeStr = "[" + prefix + "]" + getExprName(t.Elt)
	case *ast.StarExpr:
		typeStr = "*" + getExprName(t.X)
	default:
		panic("type not support")
	}
	return typeStr
}

func parsePackage(structname string) *Package {
	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.LoadSyntax,
		Tests: false,
	})
	if err != nil {
		log.Fatal(err)
	}
	pkg := pkgs[0]
	result := Package{
		Name:       pkg.Name,
		StructList: make([]Type, 0),
	}
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(node ast.Node) bool {
			decl, ok := node.(*ast.GenDecl)
			if !ok || decl.Tok != token.TYPE {
				return true
			}
			keepSearching := true
			for _, spec := range decl.Specs {
				vspec := spec.(*ast.TypeSpec)
				structType, ok := vspec.Type.(*ast.StructType)
				if !ok || structname != vspec.Name.Name {
					continue
				}
				keepSearching = false
				t := Type{
					Name:   vspec.Name.Name,
					Fields: make([]Field, 0),
				}
				for _, field := range structType.Fields.List {
					if field.Tag == nil {
						field.Tag = new(ast.BasicLit)
					}
					tags := reflect.StructTag(strings.ReplaceAll(field.Tag.Value, "`", ""))

					typeStr := getExprName(field.Type)
					for _, name := range field.Names {
						var DocList []*ast.Comment
						if field.Doc != nil {
							DocList = field.Doc.List
						}
						t.Fields = append(t.Fields, Field{
							Name:     name.Name,
							RawField: field.Type,
							DocList:  DocList,
							Tag:      tags,
							Type:     typeStr,
						})
					}
				}
				result.StructList = append(result.StructList, t)
			}
			return keepSearching
		})
	}
	return &result
}
