package views

import (
	"html/template"
	"net/http"
	"path/filepath"
)

var LAYOUTSDIR string = "views/layouts/"
var TEPLATEDIR string = "views/"
var TEMPLATEEXT string = ".gohtml"

type View struct {
	Template *template.Template
	Layout   string
}

func NewView(layout string, files ...string) *View {
	addTemplatePath(files)
	addTemplateExt(files)
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
	w.Header().Set("Content-Type", "text/html")
	switch data.(type) {
	case Data:
		// do nothing
	default:
		data = Data{
			Yield: data,
		}
	}

	return v.Template.ExecuteTemplate(w, v.Layout, data)
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := v.Render(w, nil); err != nil {
		panic(err)
	}
}

func addTemplatePath(files []string) {
	for i, f := range files {
		files[i] = TEPLATEDIR + f
	}
}

func addTemplateExt(files []string) {
	for i, f := range files {
		files[i] = f + TEMPLATEEXT
	}
}
