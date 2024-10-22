package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Chetan-gamne/bookings/internal/config"
	"github.com/Chetan-gamne/bookings/internal/forms"
	"github.com/Chetan-gamne/bookings/internal/models"
	"github.com/Chetan-gamne/bookings/internal/render"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

//  New Hanlders sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the about Home Page Handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIp := r.RemoteAddr
	m.App.Session.Put(r.Context(),"remoteIp",remoteIp)
	render.RenderTemplate(w,r,"home.page.tmpl",&models.TemplateData{})
}

// About is the About Page Handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	// Perform so logic
	stringMap := make(map[string]string)
	stringMap["test"] = "Hope, again."

	remoteIp := m.App.Session.GetString(r.Context(),"remoteIp")
	stringMap["remote_ip"] =remoteIp
	render.RenderTemplate(w,r,"about.page.tmpl",&models.TemplateData{
		StringMap: stringMap,
	})
}

// Gnerals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w,r,"generals.page.tmpl",&models.TemplateData{})
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w,r,"majors.page.tmpl",&models.TemplateData{})
}

func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w,r,"search-availability.page.tmpl",&models.TemplateData{})
}

// Post Availability
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	w.Write([]byte(fmt.Sprintf("Start date is %s and end date is %s",start,end)))
}

type jsonResponse struct {
	OK 		bool   `json:"ok"`
	Message string `json:"message"`
}

// AvailabilityJSON handles request for availability and send JSON Response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		OK: false,
		Message: "Available!",
	}

	out,err := json.MarshalIndent(resp,"","     ")
	if err != nil {
		log.Println(err)
	}

	log.Println(string(out))
	w.Header().Set("Content-Type","application/json")
	w.Write(out)
}


func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w,r,"contact.page.tmpl",&models.TemplateData{})
}
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	render.RenderTemplate(w,r,"make-reservation.page.tmpl",&models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// Handled the Posting of Reservation Form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName: r.Form.Get("last_name"),
		Email: r.Form.Get("email"),
		Phone: r.Form.Get("phone"),
	}

	form := forms.New(r.PostForm)

	// form.Has("first_name",r)
	form.Required("first_name","last_name","email","phone")
	form.MinLength("first_name",3,r)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		render.RenderTemplate(w,r,"make-reservation.page.tmpl",&models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	m.App.Session.Put(r.Context(),"reservation",reservation)

	http.Redirect(w,r,"/reservation-summary",http.StatusSeeOther)
}

func (m *Repository) ReservationSummary(w http.ResponseWriter,r *http.Request) {
	reservation,ok := m.App.Session.Get(r.Context(),"reservation").(models.Reservation);

	if !ok {
		log.Println("Cannot get item from session")
		m.App.Session.Put(r.Context(),"error","Can't get reservation from session")
		http.Redirect(w,r,"/",http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(),"reservation")
	
	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.RenderTemplate(w,r,"reservation-summary.page.tmpl",&models.TemplateData{
		Data: data,
	})
}