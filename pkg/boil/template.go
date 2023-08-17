package boil

import (
	"bytes"
	"text/template"
)

// FuncMapper can return a template.FuncMap.
type FuncMapper interface {
	FuncMap() template.FuncMap
}

// ExecuteTemplateString executes in as a text/template using data and returns it or 
// an error. If data supports FuncMapper the functions are added to the template.
func ExecuteTemplateString(in string, data any) (out string, err error) {
	var (
		tmpl = template.New("ts")
		buff = bytes.NewBuffer(nil)
	)
	if fm, ok := data.(FuncMapper); ok {
		tmpl.Funcs(fm.FuncMap())
	}
	if tmpl, err = tmpl.Parse(in); err != nil {
		return
	}
	if err = tmpl.Execute(buff, data); err != nil {
		return
	}
	return buff.String(), nil
}