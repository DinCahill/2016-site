package utils

import (
	"github.com/UniversityRadioYork/2016-site/structs"
	"html/template"
	"net/http"
	"path/filepath"
	"errors"
	"fmt"
	"github.com/bradfitz/iter"
)

// TemplatePrefix is the constant containing the filepath prefix for templates.
const TemplatePrefix = "views"

// BaseTemplates is the array of 'base' templates used in each template render.
var BaseTemplates = []string{
	"partials/header.tmpl",
	"partials/footer.tmpl",
	"elements/navbar.tmpl",
	"partials/base.tmpl",
}

// RenderTemplate renders a 2016site template on the ResponseWriter w.
//
// This function automatically adds in the 2016site base templates, performs
// error handling, and builds a global context.
//
// The PageContext context gives the context for the page to be rendered, sent
// to the template as PageContext.
// The interface{} data gives the data to be sent to the template as PageData.
//
// The string mainTmpl gives the name, relative to views, of the main
// template to render.  The variadic argument addTmpls names any additional
// templates mainTmpl depends on.
//
// RenderTemplate returns any error that occurred when rendering the template.
func RenderTemplate(w http.ResponseWriter, context structs.PageContext, data interface{}, mainTmpl string, addTmpls ...string) error {
	var err error

	td := structs.Globals{
		PageContext: context,
		PageData:    data,
	}

	ownTmpls := append(addTmpls, mainTmpl)
	baseTmpls := append(BaseTemplates, ownTmpls...)

	var tmpls []string
	for _, baseTmpl := range baseTmpls {
		tmpls = append(tmpls, filepath.Join(TemplatePrefix, baseTmpl))
	}

	t := template.New("base.tmpl")
	t.Funcs(template.FuncMap{
		"html": renderHTML,
		"N": iter.N,
		"add": func (a, b int) int {
			return a + b
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i+=2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	})
	t, err = t.ParseFiles(tmpls...)
	if err != nil {
		return err
	}

	return t.Execute(w, td)
}

// renderHTML takes some html as a string and returns a template.HTML
//
// Handles plain text gracefully.
func renderHTML(value interface{}) template.HTML {
	return template.HTML(fmt.Sprint(value))
}
