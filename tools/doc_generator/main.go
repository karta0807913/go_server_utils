package main

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"golang.org/x/tools/go/packages"
)

func main() {
	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.LoadSyntax,
		Tests: false,
	})
	if err != nil {
		log.Fatalf("parse package get error %s\n", err)
	}
	regrex, err := regexp.Compile(`^((Create|Update|Find|First).*)\.go$`)
	if err != nil {
		log.Fatalln(err)
	}
	pkg := pkgs[0]
	fileNameList := make([]string, 0)
	for _, file := range pkg.GoFiles {
		subList := regrex.FindStringSubmatch(filepath.Base(file))
		if len(subList) != 3 {
			continue
		}
		fileNameList = append(fileNameList, subList[1])
	}
	outputFile, err := os.OpenFile("main.go", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("generate main.go get error %s", err)
	}
	template.Must(template.New("").Parse(MainTemplate)).Execute(outputFile, fileNameList)
}
