package main

const MainTemplate = `package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

var (
	extraTemplate = flag.String("templates", "./_template", "template files")
)

var DocumentMap map[string]Document = map[string]Document{
    {{ range . }} "{{ . }}": {{ . }},
    {{ end }}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [option] InputFile OutputFile\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(127)
	}

	t := template.New(flag.Arg(0))
	if _, err := os.Stat(*extraTemplate); !os.IsNotExist(err) {
		err := filepath.Walk(*extraTemplate, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			file, err := os.OpenFile(path, os.O_RDONLY, 0644)
			if err != nil {
				log.Fatalf("load template %s get error %s\n", path, err)
			}
			data, err := ioutil.ReadAll(file)
			if err != nil {
				return err
			}
			t, err = t.Parse(string(data))
			if err != nil {
				log.Fatalf("parse template %s get error %s\n", path, err)
			}
			return nil
		})
		if err != nil {
			log.Fatalf("parse extra template get error %s\n", err)
		}
	}

	t.Funcs(map[string]interface{}{
		"json": func(i interface{}) string {
			data, err := json.Marshal(i)
			if err != nil {
				log.Fatalln(err)
			}
			return string(data)
		},
	})

	t, err := t.ParseFiles(flag.Arg(0))
	if err != nil {
		log.Fatalf("open file %s get error %s\n", flag.Arg(0), err)
	}

	outputFile, err := os.OpenFile(flag.Arg(1), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("open file %s get error %s\n", flag.Arg(1), err)
	}

	t.Execute(outputFile, DocumentMap)
}

type Document struct {
	Path    string
	Comment string
	Fields  []Field
}

type Field struct {
	Required bool
	Comment  string
	Name     string
	Alias    string
	Type     string
}`
