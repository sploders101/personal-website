package ht

import (
	"database/sql"
	"html/template"
	"strings"
	"time"
)

var TemplateFuncs template.FuncMap = template.FuncMap{
	"hasPrefix":      strings.HasPrefix,
	"htmlTime":       htmlTime,
	"htmlNullTime":   htmlNullTime,
	"formatTime":     formatTime,
	"formatNullTime": formatNullTime,
}

func htmlTime(ftime time.Time) string {
	return ftime.Format(time.RFC3339)
}
func htmlNullTime(ftime sql.NullTime) any {
	if ftime.Valid {
		return ftime.Time.Format(time.RFC3339)
	}
	return "null"
}

func formatTime(ftime time.Time) string {
	return ftime.Format("01/02/06 15:04 MST")
}
func formatNullTime(ftime sql.NullTime) string {
	if ftime.Valid {
		return formatTime(ftime.Time)
	}
	return "Never"
}
