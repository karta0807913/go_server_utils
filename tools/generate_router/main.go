package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
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
	required   = flag.String("require", "", "input required fields, default read gorm tag in struct which is not null without primaryKey and default")
	minItem    = flag.Uint("minItem", 1, "minimum number of items to choose")
	options    = flag.String("options", "", "input options fields")
	decoder    = flag.String("decoder", "json", "decoder: xml,json or etc")
	ignore     = flag.String("ignore", "", "which field should ignore")
	indexField = flag.String("indexField", "", "for an update index")
	max_limit  = flag.Uint("max_limit", 20, "search limit")
	min_limit  = flag.Uint("min_limit", 0, "min_limit")
)

const DocFile string = "doc.json"

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
		return
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

	_, err := os.Stat(path.Join(rootDir, DocFile))
	if os.IsNotExist(err) {
		ioutil.WriteFile(path.Join(rootDir, DocFile), []byte("{}\n"), 0642)
	}
	requireSet := NewCommaSet(*required)
	optionsSet := NewCommaSet(*options)
	ignoreSet := NewCommaSet(*ignore)

	parsedPKG := parsePackage([]string{*typeName})
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
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal("open file error ", err)
	}
	if *method == "Create" {
		temp := MethodCreate(MethodCreateParams{
			ParsedType: parsedTypes[0],
			RequireSet: requireSet,
			OptionsSet: optionsSet,
			IgnoreSet:  ignoreSet,
			TagKey:     *decoder,
			MinItem:    *minItem,
		})
		temp.Package = parsedPKG.Name
		t := template.New("").Funcs(funcMap)
		t = template.Must(t.Parse(CreateOrUpdateTemplate))
		t.Execute(file, temp)
	} else if *method == "Update" {
		temp := MethodUpdate(MethodUpdateParams{
			ParsedType: parsedTypes[0],
			RequireSet: requireSet,
			OptionsSet: optionsSet,
			IgnoreSet:  ignoreSet,
			IndexField: *indexField,
			TagKey:     *decoder,
			MinItem:    *minItem,
		})
		temp.Package = parsedPKG.Name
		t := template.New("").Funcs(funcMap)
		t = template.Must(t.Parse(CreateOrUpdateTemplate))
		t.Execute(file, temp)
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
	} else {
		log.Fatal("method not support now :<")
	}
	file.Close()
	cmd := exec.Command("go", "fmt")
	if err := cmd.Run(); err != nil {
		fmt.Println("can't find gofmt to format the code")
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
	Doc      *ast.CommentGroup
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

func parsePackage(structname []string) *Package {
	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.LoadSyntax,
		Tests: false,
	})
	if err != nil {
		log.Fatal(err)
	}
	pkg := pkgs[0]
	result := make([]Type, 0)
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(node ast.Node) bool {
			decl, ok := node.(*ast.GenDecl)
			if !ok || decl.Tok != token.TYPE {
				return true
			}
			for _, spec := range decl.Specs {
				vspec := spec.(*ast.TypeSpec)
				structType, ok := vspec.Type.(*ast.StructType)
				if !ok {
					continue
				}
				if structname != nil {
					con := false
					for _, val := range structname {
						if val == vspec.Name.Name {
							con = true
							break
						}
					}
					if !con {
						continue
					}
				}
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
						t.Fields = append(t.Fields, Field{
							Name:     name.Name,
							RawField: field.Type,
							Tag:      tags,
							Doc:      field.Doc,
							Type:     typeStr,
						})
					}
				}
				result = append(result, t)
			}
			return true
		})
	}
	return &Package{
		Name:       pkg.Name,
		StructList: result,
	}
}
