package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Chetan-gamne/bookings/internal/config"
	"github.com/Chetan-gamne/bookings/internal/handlers"
	"github.com/Chetan-gamne/bookings/internal/models"
	"github.com/Chetan-gamne/bookings/internal/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":3001"

var app config.AppConfig
var session *scs.SessionManager

func main() {
	// what am I going to put in the session
	gob.Register(models.Reservation{})

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