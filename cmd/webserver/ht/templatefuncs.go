package ht

import (
	"database/sql"
	"html/template"
	"strings"
	"time"
)

var TemplateFuncs template.FuncMap = template.FuncMap{
	"hasPrefix":      strings.HasPrefix,
	"jsTime":         jsTime,
	"jsNullTime":     jsNullTime,
	"formatTime":     formatTime,
	"formatNullTime": formatNullTime,
}

func jsTime(ftime time.Time) int64 {
	return ftime.UnixMilli()
}
func jsNullTime(ftime sql.NullTime) any {
	if ftime.Valid {
		return ftime.Time.UnixMilli()
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
