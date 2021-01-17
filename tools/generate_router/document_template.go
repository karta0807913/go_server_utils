package main

import (
	"io"
	"text/template"
)

// tr instance of SearchTemplateRoot or CreateAndUpdateTemplateRoot
func GenerateDocument(method string, tr interface{}, writer io.Writer) error {
	t := template.New("")
	t = template.Must(t.Parse(`package main

var ` + method + `{{ .StructName }} Document = Document{
    Path: "` + *apiPath + `",
    Comment: "",
	Mode: "{{ .Mode }}",
    Fields: []Field{
        {{ if or (eq .Mode "Create") (eq .Mode "Updates") }}{{ with .IndexField }}{
            Required: true,
            Comment: "{{ .Doc }}",
            Name: "{{ .Name }}",
            Alias: "{{ .Alias }}",
            Type: "{{ .Type }}",
        },{{ end }}{{ end }}{{ range .RequiredFields }}{
            Required: true,
            Comment: "{{ .Doc }}",
            Name: "{{ .Name }}",
            Alias: "{{ .Alias }}",
            Type: "{{ .Type }}",
        },{{ end }}{{ range .OptionalFields }}{
            Required: false,
            Comment: "{{ .Doc }}",
            Name: "{{ .Name }}",
            Alias: "{{ .Alias }}",
            Type: "{{ .Type }}",
        },{{ end }}
    },
}`))
	return t.Execute(writer, tr)
}
