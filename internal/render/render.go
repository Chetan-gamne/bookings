package render

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Chetan-gamne/bookings/internal/config"
	"github.com/Chetan-gamne/bookings/internal/models"
	"github.com/justinas/nosurf"
)

var functions = template.FuncMap{
	"humanDate":HumanDate,
	"formatDate":FormatDate,
	"iterate":Iterate,
	"add":Add,
}


var app *config.AppConfig

func Add(a, b int) int {
	return a+b
}

// Iterate return slice of ints, starting as 1, going to count
func Iterate(count int) []int {
	var i int
	var items []int 
	for i = 0; i < count; i++ {
		items = append(items, i)
	}
	return items
}

func NewRenderer(a *config.AppConfig) {
	app = a
}

// HumanDate return time in YYYY-MM-DD format
func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}

func AddDefaultData(td *models.TemplateData,r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(),"flash")
	td.Error = app.Session.PopString(r.Context(),"error")
	td.Warning = app.Session.PopString(r.Context(),"warning")
	td.CSRFToken = nosurf.Token(r)

	if app.Session.Exists(r.Context(),"user_id") {
		td.IsAuthenticated = 1;
	}
	return td;
}

// RenderTemplate renders templates using html/template
func Template(w http.ResponseWriter, r *http.Request,tmpl string,td *models.TemplateData) {
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
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
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
