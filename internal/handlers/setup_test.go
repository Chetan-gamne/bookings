package handlers

import (
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Chetan-gamne/bookings/internal/config"
	"github.com/Chetan-gamne/bookings/internal/driver"
	"github.com/Chetan-gamne/bookings/internal/helpers"
	"github.com/Chetan-gamne/bookings/internal/models"
	"github.com/Chetan-gamne/bookings/internal/render"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
)

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path: "/",
		Secure: false,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

// SessionLoad loads and saves the session on every request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

func getRoutes() http.Handler {
	// what am I going to put in the session
	gob.Register(models.Reservation{})


	// change this value to True in Production
	app.InProduction = false

	infoLog = log.New(os.Stdout,"INFO\t",log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout,"Error\t",log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// connect to database
	log.Println("Connecting to database...")
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=bookings user=postgres password=pspl@123")

	if err != nil {
		log.Fatal("Cannot Connect to database ! Dying...")
	}

	

	tc,err := render.CreateTemplateCache()

	if err != nil {
		log.Fatal("Cannot Create template Cache")
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := NewRepo(&app,db)
	NewHandlers(repo)

	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/",Repo.Home)
	mux.Get("/about",Repo.About)
	mux.Get("/generals-quarters",Repo.Generals)
	mux.Get("/majors-suite",Repo.Majors)

	mux.Get("/search-availability",Repo.Availability)
	mux.Post("/search-availability",Repo.PostAvailability)
	mux.Post("/search-availability-json",Repo.AvailabilityJSON)
	mux.Get("/choose-room/{id}",Repo.ChooseRoom)
	mux.Get("/book-room",Repo.BookRoom)
	
	mux.Get("/contact",Repo.Contact)
	
	mux.Get("/make-reservation",Repo.Reservation)
	mux.Post("/make-reservation",Repo.PostReservation)
	mux.Get("/reservation-summary",Repo.ReservationSummary)


	fileServer := http.FileServer(http.Dir("../../static/"))
	mux.Handle("/static/*",http.StripPrefix("/static",fileServer))

	return mux
}