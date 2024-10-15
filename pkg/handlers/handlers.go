package handlers

import (
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
	render.RenderTemplate(w,"home.html",&models.TemplateData{})
}

// About is the About Page Handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	// Perform so logic
	stringMap := make(map[string]string)
	stringMap["test"] = "Hope, again."

	remoteIp := m.App.Session.GetString(r.Context(),"remoteIp")
	stringMap["remote_ip"] =remoteIp
	render.RenderTemplate(w,"about.html",&models.TemplateData{
		StringMap: stringMap,
	})
}

