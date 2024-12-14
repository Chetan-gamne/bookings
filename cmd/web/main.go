package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Chetan-gamne/bookings/internal/config"
	"github.com/Chetan-gamne/bookings/internal/driver"
	"github.com/Chetan-gamne/bookings/internal/handlers"
	"github.com/Chetan-gamne/bookings/internal/helpers"
	"github.com/Chetan-gamne/bookings/internal/models"
	"github.com/Chetan-gamne/bookings/internal/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":3001"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger
func main() {
	db,err := run()

	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()
	defer close(app.MailChan)

	fmt.Println("Started Mail Listener")
	listenForMail()

	fmt.Println(fmt.Sprintf("Starting application on port %s",portNumber));

	srv := &http.Server{
		Addr: portNumber,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	log.Fatal(err)
}


func run () (*driver.DB , error) {

	// what am I going to put in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

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
		fmt.Println(err)
		log.Fatal("Cannot Create template Cache")
		return nil,err
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app,db)
	handlers.NewHandlers(repo)

	render.NewRenderer(&app)
	helpers.NewHelpers(&app)

	return db,nil;
}