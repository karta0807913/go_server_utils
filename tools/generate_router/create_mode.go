package main

// create mode default every fields (without primaryKey) are required

import (
	"strings"
)

type MethodCreateParams struct {
	ParsedType Type
	Doc        *Document
	RequireSet *CommaSet
	OptionsSet *CommaSet
	IgnoreSet  *CommaSet
	DefaultSet *CommaSet
	TagKey     string
}

// Create *TemplateRoot and update *Document
func MethodCreate(arg MethodCreateParams) *CreateAndUpdateTemplateRoot {
	templateRoot := CreateAndUpdateTemplateRoot{
		TemplateRoot: TemplateRoot{
			FuncName:       *method,
			StructName:     arg.ParsedType.Name,
			Decoder:        "ShouldBind" + strings.ToUpper(arg.TagKey),
			RequiredFields: make([]TemplateField, 0),
			OptionalFields: make([]TemplateField, 0),
			DefaultFields:  make([]TemplateField, 0),
			Mode:           "Create",
		},
		MinItem: *minItem,
	}

	for _, field := range arg.ParsedType.Fields {
		tf, tags, join := parseFields(templateRoot.TemplateRoot, field, arg.TagKey, arg.TagKey)
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
		} else if arg.DefaultSet.CheckAndDelete(field.Name) {
			tf.Tag = "`" + strings.Join(tags, " ") + "`"
			templateRoot.DefaultFields = append(templateRoot.DefaultFields, tf)
		} else {
			// if is primary key, pass
			if join >= 16 {
				continue
			}
			tf.Tag = "`" + strings.Join(
				append(tags, `binding:"required"`), " ") + "`"
			templateRoot.RequiredFields = append(templateRoot.RequiredFields, tf)
		}
	}
	return &templateRoot
}
