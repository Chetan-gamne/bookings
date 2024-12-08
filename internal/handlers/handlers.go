package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Chetan-gamne/bookings/internal/config"
	"github.com/Chetan-gamne/bookings/internal/driver"
	"github.com/Chetan-gamne/bookings/internal/forms"
	"github.com/Chetan-gamne/bookings/internal/helpers"
	"github.com/Chetan-gamne/bookings/internal/models"
	"github.com/Chetan-gamne/bookings/internal/render"
	"github.com/Chetan-gamne/bookings/internal/repository"
	"github.com/Chetan-gamne/bookings/internal/repository/dbrepo"
	"github.com/go-chi/chi/v5"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig,db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB: dbrepo.NewPostgresRepo(db.SQL,a),
	}
}

//  New Hanlders sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the about Home Page Handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	
	render.Template(w,r,"home.page.tmpl",&models.TemplateData{})
}

// About is the About Page Handler
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	
	render.Template(w,r,"about.page.tmpl",&models.TemplateData{})
}

// Gnerals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w,r,"generals.page.tmpl",&models.TemplateData{})
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w,r,"majors.page.tmpl",&models.TemplateData{})
}

func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w,r,"search-availability.page.tmpl",&models.TemplateData{})
}

// Post Availability
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "2006-01-02"
	startDate,err := time.Parse(layout,start)
	if err!=nil {
		helpers.ServerError(w,err)
		return
	}

	endDate,err := time.Parse(layout,end)
	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	rooms,err := m.DB.SearchAvailabilityForAllRooms(startDate,endDate)

	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	if len(rooms) == 0 {
		// no Availability
		m.App.Session.Put(r.Context(),"error","No Availabilty")
		http.Redirect(w,r,"/search-availability",http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate: endDate,
	}

	m.App.Session.Put(r.Context(),"reservation",res)

	render.Template(w,r,"choose-room.page.tmpl",&models.TemplateData{
		Data: data,
	})

}

type jsonResponse struct {
	OK 		bool   `json:"ok"`
	Message string `json:"message"`
	RoomID string  `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate string `json:"end_date"`
}

// AvailabilityJSON handles request for availability and send JSON Response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	layout := "2006-01-02"
	startDate,_ := time.Parse(layout,sd)
	endDate,_ := time.Parse(layout,ed)

	roomID,_ := strconv.Atoi(r.Form.Get("room_id"))

	available, _ := m.DB.SearchAvailabilityByDatesByRoomID(startDate,endDate,roomID)

	resp := jsonResponse{
		OK: available,
		Message: "",
		StartDate: sd,
		EndDate: ed,
		RoomID: strconv.Itoa(roomID),
	}

	out,err := json.MarshalIndent(resp,"","     ")
	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	log.Println(string(out))
	w.Header().Set("Content-Type","application/json")
	w.Write(out)
}


func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w,r,"contact.page.tmpl",&models.TemplateData{})
}
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	res,ok := m.App.Session.Get(r.Context(),"reservation").(models.Reservation)

	if !ok {
		helpers.ServerError(w,errors.New("Cannot get Reservation from session"))
		return
	}

	room, err := m.DB.GetRoomByID(res.RoomID)

	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	res.Room.RoomName = room.RoomName

	m.App.Session.Put(r.Context(), "reservation", res)

	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	fmt.Println("sd: ",sd)
	fmt.Println("ed: ",ed)

	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed 


	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w,r,"make-reservation.page.tmpl",&models.TemplateData{
		Form: forms.New(nil),
		Data: data,
		StringMap: stringMap,
	})
}

// Handled the Posting of Reservation Form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	reservation,ok := m.App.Session.Get(r.Context(),"reservation").(models.Reservation)

	if !ok {
		helpers.ServerError(w,errors.New("Can't get from Session"))
		return
	}

	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Phone = r.Form.Get("phone")
	reservation.Email = r.Form.Get("email")

	form := forms.New(r.PostForm)

	// form.Has("first_name",r)
	form.Required("first_name","last_name","email","phone")
	form.MinLength("first_name",3,r)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		render.Template(w,r,"make-reservation.page.tmpl",&models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID,err := m.DB.InsertReservation(reservation)

	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	m.App.Session.Put(r.Context(),"reservation",reservation)

	restriction := models.RoomRestriction{
		StartDate: reservation.StartDate,
		EndDate: reservation.EndDate,
		RoomID: reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 3,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	m.App.Session.Put(r.Context(),"reservation",reservation)

	http.Redirect(w,r,"/reservation-summary",http.StatusSeeOther)
}

// display summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter,r *http.Request) {
	reservation,ok := m.App.Session.Get(r.Context(),"reservation").(models.Reservation);

	if !ok {
		m.App.ErrorLog.Println("can't get error from session")
		m.App.Session.Put(r.Context(),"error","Can't get reservation from session")
		http.Redirect(w,r,"/",http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(),"reservation")
	
	data := make(map[string]interface{})
	data["reservation"] = reservation

	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.Template(w,r,"reservation-summary.page.tmpl",&models.TemplateData{
		Data: data,
		StringMap: stringMap,
	})
}

// Displays available rooms
func (m *Repository) ChooseRoom(w http.ResponseWriter,r *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(r,"id"))

	if err != nil {
		helpers.ServerError(w,err)
		return
	}
	m.App.Session.Get(r.Context(),"reservation")

	res, ok := m.App.Session.Get(r.Context(),"reservation").(models.Reservation)

	if !ok {
		helpers.ServerError(w,err)
		return
	}

	res.RoomID = roomID

	m.App.Session.Put(r.Context(),"reservation",res)

	http.Redirect(w,r,"/make-reservation",http.StatusSeeOther)

}

// BookRoom takes URL Parameters Builds sessional variable, and takes user to reservation screen
func (m *Repository) BookRoom(w http.ResponseWriter,r *http.Request) {

	roomID,_ := strconv.Atoi(r.URL.Query().Get("id"));
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	layout := "2006-01-02"
	startDate,_ := time.Parse(layout,sd)
	endDate,_ := time.Parse(layout,ed)

	var res models.Reservation

	
	room, err := m.DB.GetRoomByID(roomID)
	
	if err != nil {
		helpers.ServerError(w,err)
		return
	}
	
	res.Room.RoomName = room.RoomName
	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate

	m.App.Session.Put(r.Context(),"reservation",res)


	http.Redirect(w,r,"make-reservation",http.StatusSeeOther)
}