package render

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/Chetan-gamne/bookings/pkg/config"
	"github.com/Chetan-gamne/bookings/pkg/models"
)

var app *config.AppConfig

func NewTemplates(a *config.AppConfig) {
	app = a
}


func AddDefaultData(td *models.TemplateData) *models.TemplateData {

	return td;
}

// RenderTemplate renders templates using html/template
func RenderTemplate(w http.ResponseWriter, tmpl string,td *models.TemplateData) {
	// Create a Template Cache
	// get the Template Cache From the app Config
	var tc map[string]*template.Template
	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc,_ = CreateTemplateCache()
	}
	t, ok := tc[tmpl]

	if !ok {
		log.Fatal("Could Not get Template From Template Cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td)
	_ = t.Execute(buf,td)

	// Render the Template

	_,err := buf.WriteTo(w)

	if err != nil {
		log.Println(err)
	}
}


func CreateTemplateCache() (map[string]*template.Template,error) {
	myCache := map[string]*template.Template{}

	// get all of the files named *.html from ./templates

	pages,err := filepath.Glob("../../templates/*.html")

	if err != nil {
		return myCache,err
	}

	// range through all files ending with .html

	for _,page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).ParseFiles(page)
		if err != nil {
			return myCache,err
		}

		matches, err := filepath.Glob("../../templates/base.html")

		if err != nil {
			return myCache,err
		}

		if len(matches) > 0 {
			ts,err = ts.ParseGlob("../../templates/base.html")
			if err != nil {
				return myCache,err
			}
		}

		myCache[name] = ts
	}

	return myCache,nil
} 
