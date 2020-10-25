package main

import (
	"strings"
)

type MethodSearchParams struct {
	ParsedType Type
	Doc        *Document
	RequireSet *CommaSet
	OptionsSet *CommaSet
	IgnoreSet  *CommaSet
	TagKey     string
}

// default the "index" will be join as options
func MethodSearch(arg MethodSearchParams) *SearchTemplateRoot {
	templateRoot := SearchTemplateRoot{
		TemplateRoot: TemplateRoot{
			FuncName:       *method,
			StructName:     arg.ParsedType.Name,
			Decoder:        "ShouldBindQuery",
			RequiredFields: make([]TemplateField, 0),
			OptionalFields: make([]TemplateField, 0),
		},
	}

	if max_limit == nil {
		templateRoot.MaxLimit = 20
	} else {
		templateRoot.MaxLimit = *max_limit
	}

	if min_limit == nil {
		templateRoot.MinLimit = 0
	} else {
		templateRoot.MinLimit = *min_limit
	}

	for _, field := range arg.ParsedType.Fields {
		tf, tags, flag := parseFields(field, arg.TagKey, "form")

		// if this field is required
		if arg.IgnoreSet.CheckAndDelete(field.Name) {
			continue
		} else if arg.RequireSet.CheckAndDelete(field.Name) {
			tf.Tag = "`" + strings.Join(
				append(tags, `binding:"required"`), " ") + "`"
			templateRoot.RequiredFields = append(templateRoot.RequiredFields, tf)
		} else if arg.OptionsSet.CheckAndDelete(field.Name) {
			tf.Tag = "`" + strings.Join(tags, " ") + "`"
			templateRoot.OptionalFields = append(templateRoot.OptionalFields, tf)
		} else {
			if flag < 4 {
				continue
			}
			tf.Tag = "`" + strings.Join(tags, " ") + "`"
			templateRoot.OptionalFields = append(templateRoot.OptionalFields, tf)
		}
	}
	return &templateRoot
}
