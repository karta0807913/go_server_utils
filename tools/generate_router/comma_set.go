package main

import (
	"strings"
)

type CommaSet map[string]bool

func NewCommaSet(arg string) *CommaSet {
	commaSet := CommaSet(make(map[string]bool))
	for _, key := range strings.Split(arg, ",") {
		commaSet[key] = true
	}
	return &commaSet
}

func (cm *CommaSet) CheckAndDelete(key string) bool {
	defer delete(*cm, key)
	return (*cm)[key]
}

func (cm CommaSet) ToArray() []string {
	result := make([]string, 0)
	for key, _ := range cm {
		result = append(result, key)
	}
	return result
}
