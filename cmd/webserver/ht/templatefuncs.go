package ht

import (
	"html/template"
	"strings"
)

var TemplateFuncs template.FuncMap = template.FuncMap{
	"hasPrefix": strings.HasPrefix,
}
