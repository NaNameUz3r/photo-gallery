package views

import (
	"html/template"
	"net/http"
	"path/filepath"
)

var LAYOUTSDIR string = "views/layouts/"
var TEMPLATEEXT string = ".gohtml"

type View struct {
	Template *template.Template
	Layout   string
}

func NewView(layout string, files ...string) *View {
	files = append(files, parseLayoutFiles()...)

	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
		Layout:   layout,
	}
}

func parseLayoutFiles() []string {
	files, err := filepath.Glob(LAYOUTSDIR + "*" + TEMPLATEEXT)
	if err != nil {
		panic(err)
	}
	return files
}

func (v *View) Render(w http.ResponseWriter, data interface{}) error {
	return v.Template.ExecuteTemplate(w, v.Layout, data)
}
