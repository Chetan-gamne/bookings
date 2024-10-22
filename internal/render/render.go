package render

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/Chetan-gamne/bookings/internal/config"
	"github.com/Chetan-gamne/bookings/internal/models"
	"github.com/justinas/nosurf"
)

var app *config.AppConfig

func NewTemplates(a *config.AppConfig) {
	app = a
}


func AddDefaultData(td *models.TemplateData,r *http.Request) *models.TemplateData {
	td.CSRFToken = nosurf.Token(r)
	return td;
}

// RenderTemplate renders templates using html/template
func RenderTemplate(w http.ResponseWriter, r *http.Request,tmpl string,td *models.TemplateData) {
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

	td = AddDefaultData(td,r)
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

	pages,err := filepath.Glob("../../templates/*.page.tmpl")

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

		matches, err := filepath.Glob("../../templates/*.layout.tmpl")

		if err != nil {
			return myCache,err
		}

		if len(matches) > 0 {
			ts,err = ts.ParseGlob("../../templates/*.layout.tmpl")
			if err != nil {
				return myCache,err
			}
		}

		myCache[name] = ts
	}

	return myCache,nil
} 
