package views

import (
	"html/template"
	"io"
	"path/filepath"
)

// TODO Rename directory to "views"

const (
	// LayoutDir  directory containing the template layouts
	LayoutDir = "views/layouts/"

	// TemplateExt  file extension of the templates
	TemplateExt = ".gohtml"
)

// View template and layout combination representing a web page
type View struct {
	Layout   string
	Template *template.Template
}

// New constructs a new view
func New(layout string, files ...string) (v *View) {
	files = append(files, layoutFiles()...)
	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}
	return &View{
		Layout:   layout,
		Template: t,
	}
}

func layoutFiles() (files []string) {
	files, _ = filepath.Glob(LayoutDir + "*" + TemplateExt)
	return
}

// Render renders a template to a writer, using the given data
func (v *View) Render(w io.Writer, data interface{}) error {
	return v.Template.ExecuteTemplate(w, v.Layout, data)
}
