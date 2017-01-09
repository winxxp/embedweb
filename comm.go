package embedweb

import (
	"net/http"
	"html/template"
)

type X map[string]interface{}

func toHtml(w http.ResponseWriter, tmpl string, data X) error {
	t, err := template.New("log").Parse(tmpl)
	if err != nil {
		return err
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return t.Execute(w, data)
}