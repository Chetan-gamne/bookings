package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Chetan-gamne/bookings/pkg/config"
	"github.com/Chetan-gamne/bookings/pkg/handlers"
	"github.com/Chetan-gamne/bookings/pkg/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":3001"

var app config.AppConfig
var session *scs.SessionManager

func main() {

	// change this value to True in Production
	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	tc,err := render.CreateTemplateCache()

	if err != nil {
		fmt.Println(err)
		log.Fatal("Cannot Create template Cache")
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	render.NewTemplates(&app)
	

	// http.HandleFunc("/",handlers.Repo.Home)
	// http.HandleFunc("/about",handlers.Repo.About)

	fmt.Println(fmt.Sprintf("Starting application on port %s",portNumber));

	srv := &http.Server{
		Addr: portNumber,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	log.Fatal(err)
}