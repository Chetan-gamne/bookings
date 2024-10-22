package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Chetan-gamne/bookings/pkg/config"
	"github.com/Chetan-gamne/bookings/pkg/models"
	"github.com/Chetan-gamne/bookings/pkg/render"
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
	render.RenderTemplate(w,r,"make-reservation.page.tmpl",&models.TemplateData{})
}



