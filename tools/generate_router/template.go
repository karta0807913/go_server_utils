package main

import (
	"fmt"
	"go/ast"
	"strings"
)

const CreateOrUpdateTemplate = `
package {{ .Package }}

{{ $TableName := (tablename .StructName) }}

// this file generate by go generate, please don't edit it
// data will put into struct
func (insert *{{ .StructName }}){{ .FuncName }}(c *gin.Context, db *gorm.DB) error {
    type Body struct {
      {{ with .IndexField }}{{ .Name }} {{ .Type }} {{ .Tag }} {{ end }}
      {{ range .RequiredFields }}{{ .Name }}  {{ remove_pointer_type . }} {{ .Tag }}
      {{ end }}
      {{ range .OptionalFields }}{{ .Name }} {{ to_pointer_type . }} {{ .Tag }}
      {{ end }}
    }
    var body Body;
    err := c.{{ .Decoder }}(&body)
    if err != nil {
        return err
    }
    {{ with .IndexField }}insert.{{ .Name }} = body.{{ .Name }}{{ end }}

    {{/* select array */}}
    {{ if or (ne (len .RequiredFields) 0) (ne (len .DefaultFields) 0)}}
      selectField := []string {
        {{ range .RequiredFields }}"{{ .Column }}",
        {{ end }}{{ range .DefaultFields }}"{{ .Column }}",
        {{ end }}
      }
    {{ else }}
      selectField := make([]string, 0)
    {{ end }}

    {{/* put options */}}
    {{ range .OptionalFields }}
      if body.{{ .Name }} != nil {
        selectField = append(selectField, "{{ .Column }}")
        insert.{{ .Name }} = {{ if .StarExpr }}{{ else }}*{{ end }}body.{{ .Name }}
      }
    {{ end }}

    {{/* check options */}}
    {{ if ne .MinItem 0 }}
      {{ if ne (len .OptionalFields) 0}}
        if len(selectField) < ({{ len .RequiredFields }} + {{ len .DefaultFields }} + {{ .MinItem }}) {
          return errors.New("require at least one option")
        }
      {{ end }}
    {{ end }}


    {{ range .RequiredFields }}insert.{{ .Name }} = {{ if .StarExpr }}&{{ else }}{{ end }}body.{{ .Name }}
    {{ end }}

    {{/* create or update */}}
    return db.Select(
    selectField[0], selectField[1:],
    ){{ with .IndexField }}.Where("{{ $TableName }}.{{ .Column }}=?", body.{{.Name}}){{ end }}.{{ .Mode }}(&insert).Error
}
`

const FirstTemplate = `
package {{ .Package }}

{{ $TableName := (tablename .StructName) }}

// this file generate by go generate, please don't edit it
// search options will put into struct
func (item *{{ .StructName }}) {{ .FuncName }}(c *gin.Context, db *gorm.DB) error {
    type Body struct {
      {{ range .RequiredFields }}{{ .Name }}  {{ .Type }} {{ .Tag }}
      {{ end }}
      {{ range .OptionalFields }}{{ .Name }} *{{ .Type }} {{ .Tag }}
      {{ end }}
    }

    var body Body;
    err := c.{{ .Decoder }}(&body)
    if err != nil {
        return err
    }
    {{/* if decode success, search the specific data */}}
    {{ if or (ne (len .RequiredFields) 0) (ne (len .DefaultFields) 0)}}
      whereField := []string {
        {{ range .RequiredFields }}"{{ $TableName }}.{{ .Column }}=?",
        {{ end }}{{ range .DefaultFields }}"{{ $TableName }}{{ .Column }}",
        {{ end }}
      }
      valueField := []interface{}{
        {{ range .RequiredFields }}body.{{ .Name }},
        {{ end }}{{ range .DefaultFields }}item.{{ .Name }},
        {{ end }}
      }
      {{ range .RequiredFields }}
        item.{{ .Name }} = body.{{ .Name }}
      {{ end }}
    {{ else }}
      whereField := make([]string, 0)
      valueField := make([]interface{}, 0)
    {{ end }}

    {{/* put options */}}
    {{ range .OptionalFields }}
      if body.{{ .Name }} != nil {
        whereField = append(whereField, "{{ $TableName }}.{{ .Column }}=?")
        valueField = append(valueField, body.{{ .Name }})
        item.{{ .Name }} = *body.{{ .Name }}
      }
    {{ end }}

    {{ if eq (len .RequiredFields) 0}}
      if len(valueField) < ({{ len .RequiredFields }} + {{ len .DefaultFields }} + 1) {
        return errors.New("require option")
      }
    {{ end }}

    {{/* return item */}}
    err = db.Where(
      strings.Join(whereField, " and "),
      valueField[0], valueField[1:],
    ).First(item).Error
    return err
}`

const FindTemplate = `
package {{ .Package }}

{{ $TableName := (tablename .StructName) }}

// this file generate by go generate, please don't edit it
// search options will put into struct
func (item *{{ .StructName }}) {{ .FuncName }}(c *gin.Context, db *gorm.DB) ([]{{ .StructName }}, error) {
    type Body struct {
      {{ range .RequiredFields }}{{ .Name }}  {{ .Type }} {{ .Tag }}
      {{ end }}
      {{ range .OptionalFields }}{{ .Name }} *{{ .Type }} {{ .Tag }}
      {{ end }}
    }
    var body Body;
    var err error;
    _ = c.{{ .Decoder }}(&body)

    {{/* if decode success, search the specific data */}}
    {{ if or (ne (len .RequiredFields) 0) (ne (len .DefaultFields) 0)}}
      whereField := []string {
        {{ range .RequiredFields }}"{{ $TableName }}.{{ .Column }}=?",
        {{ end }}{{ range .DefaultFields }}"{{ $TableName }}.{{ .Column }}=?",
        {{ end }}
      }
      valueField := []interface{}{
        {{ range .RequiredFields }}body.{{ .Name }},
        {{ end }}{{ range .DefaultFields }}item.{{ .Name }},
        {{ end }}
      }
      {{ range .RequiredFields }}
      item.{{ .Name }} = body.{{ .Name }}{{ end }}
    {{ else }}
      whereField := make([]string, 0)
      valueField := make([]interface{}, 0)
    {{ end }}

    {{/* put options */}}
    {{ range .OptionalFields }}
      if body.{{ .Name }} != nil {
        whereField = append(whereField, "{{ $TableName }}.{{ .Column }}=?")
        valueField = append(valueField, body.{{ .Name }})
        item.{{ .Name }} = *body.{{ .Name }}
      }
    {{ end }}

    {{/* return item */}}
     var limit int = {{ .MaxLimit }}
     slimit, ok := c.GetQuery("limit")
     if ok {
       limit, err = strconv.Atoi(slimit)
       if err != nil {
         limit = {{ .MaxLimit }}
       } else {
         if limit <= {{ .MinLimit }} || {{ .MaxLimit }} < limit {
           limit = {{ .MaxLimit }}
         }
       }
    }
    soffset, ok := c.GetQuery("offset")
	var offset int
    if ok {
		offset, err = strconv.Atoi(soffset)
		if err != nil {
			offset = 0
		} else if offset < 0 {
			offset = 0
		}
	} else {
		offset = 0
	}
    var result []{{ .StructName }}
    if len(whereField) != 0 {
	  db = db.Where(
        strings.Join(whereField, " and "),
        valueField[0], valueField[1:],
      )
    }
	err = db.Limit(limit).Offset(offset).Find(&result).Error
    return result, err
}
`

type TemplateRoot struct {
	Package        string
	FuncName       string
	StructName     string
	Decoder        string
	Mode           string
	RequiredFields []TemplateField
	OptionalFields []TemplateField
	DefaultFields  []TemplateField
}

type SearchTemplateRoot struct {
	TemplateRoot
	MaxLimit uint
	MinLimit uint
}

type CreateAndUpdateTemplateRoot struct {
	TemplateRoot
	IndexField *TemplateField
	MinItem    uint
}

type TemplateField struct {
	Name     string
	Type     string
	Tag      string
	Doc      string
	Column   string
	Alias    string
	StarExpr bool
}

func parseFields(root TemplateRoot, field Field, tagKey string, encodeKey string) (TemplateField, []string, uint8) {
	tf := TemplateField{
		Name: field.Name,
		Type: field.Type,
	}
	if field.DocList != nil && len(field.DocList) > 0 {
		tf.Doc = strings.ReplaceAll(field.DocList[0].Text, "\n", " ")
		tf.Doc = strings.Replace(tf.Doc, "//", "", 1)
	}
	tags := make([]string, 0)
	decoder, ok := field.Tag.Lookup(tagKey)
	if ok {
		tags = append(tags, fmt.Sprintf(`%s:"%s"`, encodeKey, decoder))
		tf.Alias = decoder
	} else {
		tf.Alias = field.Name
	}
	column, ok := field.Tag.Lookup("column")
	if ok {
		tf.Column = column
	} else {
		tf.Column = namer.ColumnName(root.StructName, field.Name)
	}

	_, tf.StarExpr = field.RawField.(*ast.StarExpr)

	gormTag, ok := field.Tag.Lookup("gorm")
	//     16       8      4      2        1
	// primaryKey unique index not_null default
	//     0        0      0      0        0
	var flag uint8 = 0
	if ok {
		opt := gormTag
		if strings.Contains(opt, "not null") {
			flag |= 2
		}
		if strings.Contains(opt, "primaryKey") {
			flag |= 16
		}
		if strings.Contains(opt, "index") {
			flag |= 4
		}
		if strings.Contains(opt, "unique") {
			flag |= 8
		}
	}
	return tf, tags, flag
}
