package views

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"photo-gallery/context"

	"github.com/gorilla/csrf"
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
	t, err := template.New("").Funcs(template.FuncMap{
		"csrfField": func() (template.HTML, error) {
			return "", errors.New("csrffield is not implemented")
		},
	}).ParseFiles(files...)
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

func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	var vd Data
	switch d := data.(type) {
	case Data:
		vd = d
	default:
		vd = Data{
			Yield: data,
		}
	}

	if alert := getAlert(*r); alert != nil && vd.Alert == nil {
		vd.Alert = alert
		clearAlert(w)
	}

	vd.User = context.User(r.Context())
	var buffer bytes.Buffer

	csrfField := csrf.TemplateField(r)
	tpl := v.Template.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrfField
		},
	})
	if err := tpl.ExecuteTemplate(&buffer, v.Layout, vd); err != nil {
		log.Println(err)
		http.Error(w, "Somthing went wrong. If the problem persists, please contact us", http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buffer)
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
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
