package main

import (
	"strings"
)

type MethodUpdateParams struct {
	ParsedType Type
	Doc        *Document
	RequireSet *CommaSet
	OptionsSet *CommaSet
	IgnoreSet  *CommaSet
	DefaultSet *CommaSet
	IndexField string
	TagKey     string
}

func MethodUpdate(arg MethodUpdateParams) *CreateAndUpdateTemplateRoot {
	templateRoot := CreateAndUpdateTemplateRoot{
		TemplateRoot: TemplateRoot{
			FuncName:       *method,
			StructName:     arg.ParsedType.Name,
			Decoder:        "ShouldBind" + strings.ToUpper(arg.TagKey),
			RequiredFields: make([]TemplateField, 0),
			OptionalFields: make([]TemplateField, 0),
			DefaultFields:  make([]TemplateField, 0),
			Mode:           "Updates",
		},
		MinItem: *minItem,
	}

	var indexFlag uint8 = 0
	var indexTags []string = make([]string, 0)
	for _, field := range arg.ParsedType.Fields {
		tf, tags, flag := parseFields(templateRoot.TemplateRoot, field, arg.TagKey, arg.TagKey)

		// if this field is required
		if arg.IndexField == field.Name {
			if templateRoot.IndexField != nil {
				templateRoot.IndexField.Tag = "`" + strings.Join(indexTags, " ") + "`"
				templateRoot.OptionalFields = append(templateRoot.OptionalFields,
					*templateRoot.IndexField)
			}
			templateRoot.IndexField = &tf
			indexTags = tags
			indexFlag = 31
		} else if arg.IgnoreSet.CheckAndDelete(field.Name) {
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
		} else if flag > indexFlag {
			if templateRoot.IndexField != nil {
				templateRoot.IndexField.Tag = "`" + strings.Join(indexTags, " ") + "`"
				templateRoot.OptionalFields = append(templateRoot.OptionalFields,
					*templateRoot.IndexField)
			}
			templateRoot.IndexField = &tf
			indexFlag = flag
			indexTags = tags
		} else {
			if flag >= 16 {
				continue
			}
			tf.Tag = "`" + strings.Join(tags, " ") + "`"
			templateRoot.OptionalFields = append(templateRoot.OptionalFields, tf)
		}
	}
	if templateRoot.IndexField != nil {
		templateRoot.IndexField.Tag = "`" + strings.Join(
			append(indexTags, `binding:"required"`), " ") + "`"
	}
	return &templateRoot
}
