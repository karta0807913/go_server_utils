package main

import (
	"fmt"
	"strings"
)

const CreateOrUpdateTemplate = `
package {{ .Package }}

{{ $StructName := (print .StructName "s") }}

// this file generate by go generate, please don't edit it
// data will put into struct
func (insert *{{ .StructName }}){{ .FuncName }}(c *gin.Context, db *gorm.DB) error {
    type Body struct {
      {{ with .IndexField }}{{ .Name }} {{ .Type }} {{ .Tag }} {{ end }}
      {{ range .RequiredFields }}{{ .Name }}  {{ .Type}} {{ .Tag }}
      {{ end }}
      {{ range .OptionalFields }}{{ .Name }} *{{ .Type}} {{ .Tag }}
      {{ end }}
    }
    var body Body;
    err := c.{{ .Decoder }}(&body)
    if err != nil {
        return err
    }
    {{ with .IndexField }}insert.{{ .Name }} = body.{{ .Name }}{{ end }}

    {{/* select array */}}
    {{ if ne (len .RequiredFields) 0}}
      selectField := []string {
        {{ range .RequiredFields }}"{{ .Column }}",
        {{ end }}
      }
    {{ else }}
      selectField := make([]string, 0)
    {{ end }}

    {{/* put options */}}
    {{ range .OptionalFields }}
      if body.{{ .Name }} != nil {
        selectField = append(selectField, "{{ .Column }}")
        insert.{{ .Name }} = *body.{{ .Name }}
      }
    {{ end }}

    {{/* check options */}}
    {{ if ne (len .OptionalFields) 0}}
      if len(selectField) == {{ len .RequiredFields }} {
        return errors.New("rqeuire at least one option")
      }
    {{ end }}


    {{ range .RequiredFields }}insert.{{ .Name }} = body.{{ .Name }}
    {{ end }}

    {{/* create or update */}}
    return db.Select(
    selectField[0], selectField[1:],
    ){{ with .IndexField }}.Where("{{ underscore $StructName }}.{{ .Column }}=?", body.{{.Name}}){{ end }}.{{ .Mode }}(&insert).Error
}
`

const FirstTemplate = `
package {{ .Package }}

{{ $StructName := (print .StructName "s") }}

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
    {{ if ne (len .RequiredFields) 0}}
      whereField := []string {
        {{ range .RequiredFields }}"{{ underscore $StructName }}.{{ .Column }}=?",
        {{ end }}
      }
      valueField := []interface{}{
        {{ range .RequiredFields }}body.{{ .Name }},
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
        whereField = append(whereField, "{{ underscore $StructName }}.{{ .Column }}=?")
        valueField = append(valueField, body.{{ .Name }})
        item.{{ .Name }} = *body.{{ .Name }}
      }
    {{ end }}

    {{ if eq (len .RequiredFields) 0}}
      if len(valueField) == 0 {
        return errors.New("require at least one option")
      }
    {{ end }}

    {{/* return item */}}
    err = db.Where(
      strings.Join(whereField, "and"),
      valueField,
    ).First(item).Error
    return err
}`

const FindTemplate = `
package {{ .Package }}

{{ $StructName := (print .StructName "s") }}

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
    err := c.{{ .Decoder }}(&body)

    {{/* if decode success, search the specific data */}}
    {{ if ne (len .RequiredFields) 0}}
      whereField := []string {
        {{ range .RequiredFields }}"{{ underscore $StructName }}.{{ .Column }}=?",
        {{ end }}
      }
      valueField := []interface{}{
        {{ range .RequiredFields }}body.{{ .Name }},
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
        whereField = append(whereField, "{{ underscore $StructName }}.{{ .Column }}=?")
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
    offset, err := strconv.Atoi(soffset)
    if err != nil {
      offset = 0
    } else if offset < 0 {
      offset = 0
    }
    var result []{{ .StructName }}
    if len(whereField) != 0 {
	  db = db.Where(
        strings.Join(whereField, "and"),
        valueField,
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
	IndexField     *TemplateField
	RequiredFields []TemplateField
	OptionalFields []TemplateField
}

type SearchTemplateRoot struct {
	TemplateRoot
	MaxLimit uint
	MinLimit uint
}

type TemplateField struct {
	Name   string
	Type   string
	Tag    string
	Column string
}

func parseFields(field Field, tagKey string, encodeKey string) (TemplateField, []string, uint8) {
	tf := TemplateField{
		Name: field.Name,
		Type: field.Type,
	}
	tags := make([]string, 0)
	decoder, ok := field.Tag.Lookup(tagKey)
	if ok {
		tags = append(tags, fmt.Sprintf(`%s:"%s"`, encodeKey, decoder))
	}
	column, ok := field.Tag.Lookup("column")
	if ok {
		tf.Column = column
	} else {
		tf.Column = underscore(field.Name)
	}
	gormTag, ok := field.Tag.Lookup("gorm")
	//     16       8      4      2        1
	// primaryKey unique index not_null default
	//     0        0      0      0        0
	var flag uint8 = 0
	if ok {
		opt := gormTag
		if strings.Index(opt, "not null") != -1 {
			flag |= 2
		}
		if strings.Index(opt, "primaryKey") != -1 {
			flag |= 16
		}
		if strings.Index(opt, "index") != -1 {
			flag |= 4
		}
		if strings.Index(opt, "unique") != -1 {
			flag |= 8
		}
	}
	return tf, tags, flag
}
