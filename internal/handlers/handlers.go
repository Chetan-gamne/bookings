package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

// NewTestRepo creates a new repository
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingsRepo(a),
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
	form.MinLength("first_name",3)
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
		RestrictionID: 1,
	}

	err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	// send notification
	htmlMessage := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong><br>
		Dear %s:, <br>

		This is Confirmation For Your Reservation from %s to %s.
	`,reservation.FirstName,reservation.StartDate.Format("2006-01-02"),reservation.EndDate.Format("2006-01-02"))


	msg := models.MailData{
		To: reservation.Email,
		From: "me@here.com",
		Subject: "Reservation Confirmation",
		Content: htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg


	// send notification to property owner

	htmlMessage = fmt.Sprintf((`
		<strong>Reservation Notification</strong><br>
		A reservation has been made for %s from %s to %s.
	`),reservation.Room.RoomName,reservation.StartDate.Format("2006-01-02"),reservation.EndDate.Format("2006-01-02"))

	msg = models.MailData{
		To: "me@here.com",
		From: "me@here.com",
		Subject: "Reservation Notification",
		Content: htmlMessage,
		Template: "basic.html",
	}

	m.App.MailChan <- msg

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

func (m *Repository) ShowLogin(w http.ResponseWriter,r *http.Request) {
	render.Template(w,r,"login.page.tmpl",&models.TemplateData{
		Form: forms.New(nil),
	})
}


// PostShowLogin handles logging the user in
func (m *Repository) PostShowLogin(w http.ResponseWriter,r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	var email string 
	var password string 

	email = r.Form.Get("email")
	password = r.Form.Get("password")

	form := forms.New(r.PostForm)

	form.Required("email","password")

	if !form.Valid() {
		render.Template(w,r,"login.page.tmpl",&models.TemplateData{
			Form: form,
		})
		return
	}


	id, _, err := m.DB.Authenticate(email,password)

	if err != nil {
		log.Println(err)

		m.App.Session.Put(r.Context(),"error","Invalid login credentials")
		http.Redirect(w,r,"/user/login",http.StatusSeeOther)
		return 
	}

	m.App.Session.Put(r.Context(),"user_id",id)
	m.App.Session.Put(r.Context(),"flash","Logged in successfully")
	http.Redirect(w,r,"/",http.StatusSeeOther)

}

// Logout Admin From Session
func (m *Repository) Logout(w http.ResponseWriter,r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w,r,"/user/login",http.StatusSeeOther)
}


// redirect to Admin Dashboard
func (m *Repository) AdminDashboard(w http.ResponseWriter,r *http.Request) {
	render.Template(w,r,"admin-dashboard.page.tmpl",&models.TemplateData{})
}


// Admin All Reservatios show all reservations in Admin tool
func (m *Repository) AdminAllReservations(w http.ResponseWriter,r *http.Request) {
	reservations,err := m.DB.AllReservations()

	if err != nil {
		helpers.ServerError(w,err)
		return
	}


	data := make(map[string]interface{})
	data["reservations"] = reservations


	render.Template(w,r,"admin-all-reservations.page.tmpl",&models.TemplateData{
		Data: data,
	})
}

// AdminNewReservations shows all new reservations in admin tool
func (m *Repository) AdminNewReservations(w http.ResponseWriter,r *http.Request) {
	reservations,err := m.DB.AllNewReservations()

	if err != nil {
		helpers.ServerError(w,err)
		return
	}


	data := make(map[string]interface{})
	data["reservations"] = reservations
	render.Template(w,r,"admin-new-reservations.page.tmpl",&models.TemplateData{
		Data: data,
	})
}

func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI,"/")
	id,err := strconv.Atoi(exploded[4])

	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	src := exploded[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src 

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	stringMap["month"] = month
	stringMap["year"] = year


	res,err := m.DB.GetReservationByID(id)

	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res
	

	render.Template(w,r,"admin-reservations-show.page.tmpl",&models.TemplateData{
		StringMap: stringMap,
		Data: data,
		Form: forms.New(nil),
	})
}

func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	exploded := strings.Split(r.RequestURI,"/")
	id,err := strconv.Atoi(exploded[4])

	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	src := exploded[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src


	res, err := m.DB.GetReservationByID(id)

	if err != nil {
		helpers.ServerError(w,err)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")


	err = m.DB.UpdateReservation(res)


	if err != nil {
		helpers.ServerError(w,err)
		return
	}


	month := r.Form.Get("month")
	year := r.Form.Get("year")

	m.App.Session.Put(r.Context(),"flash","Changes Saved")

	if year == "" {
		http.Redirect(w,r,fmt.Sprintf("/admin/reservations-%s",src),http.StatusSeeOther)
	} else {
		http.Redirect(w,r,fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s",year,month),http.StatusSeeOther)
	}

	http.Redirect(w,r,fmt.Sprintf("/admin/reservations-%s",src),http.StatusSeeOther)
}

// AdminReservationsCalendar displays the reservation calendar
func (m *Repository) AdminReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	// assume that there is no month/year specified
	now := time.Now()

	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	// get the first and last days of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data["rooms"] = rooms

	for _, x := range rooms {
		// create maps
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0
		}

		// get all the restrictions for the current room
		restrictions, err := m.DB.GetRestrictionsForRoomByDate(x.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		for _, y := range restrictions {
			if y.ReservationID > 0 {
				// it's a reservation
				for d := y.StartDate; d.After(y.EndDate) == false; d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-2")] = y.ReservationID
				}
			} else {
				// it's a block
				blockMap[y.StartDate.Format("2006-01-2")] = y.ID
			}
		}
		data[fmt.Sprintf("reservation_map_%d", x.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.ID)] = blockMap

		m.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", x.ID), blockMap)
	}

	render.Template(w, r, "admin-reservations-calendar.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}


// AdminProcessReservation - Marks Reservation as Process
func (m *Repository) AdminProcessReservation(w http.ResponseWriter,r *http.Request) {
	id, _  := strconv.Atoi(chi.URLParam(r,"id"))
	src := chi.URLParam(r,"src")

	_ = m.DB.UpdateProcessedForReservation(id,1)

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(),"flash","Reservation Marked as Processed")


	if year == "" {
		http.Redirect(w,r,fmt.Sprintf("/admin/reservations-%s",src),http.StatusSeeOther)
	} else {
		http.Redirect(w,r,fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s",year,month),http.StatusSeeOther)
	}

}


// AdminDeleteReservation - Deletes Reservation 
func (m *Repository) AdminDeleteReservation(w http.ResponseWriter,r *http.Request) {
	id, _  := strconv.Atoi(chi.URLParam(r,"id"))
	src := chi.URLParam(r,"src")

	_ = m.DB.DeleteReservation(id)

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	m.App.Session.Put(r.Context(),"flash","Reservation Deleted")

	if year == "" {
		http.Redirect(w,r,fmt.Sprintf("/admin/reservations-%s",src),http.StatusSeeOther)
	} else {
		http.Redirect(w,r,fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s",year,month),http.StatusSeeOther)
	}
}

// AdminPostReservationsCalendar handles post of reservation calendar
func (m *Repository) AdminPostReservationsCalendar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	// process blocks
	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := forms.New(r.PostForm)

	for _, x := range rooms {
		// Get the block map from the session. Loop through entire map, if we have an entry in the map
		// that does not exist in our posted data, and if the restriction id > 0, then it is a block we need to
		// remove.
		curMap := m.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", x.ID)).(map[string]int)
		for name, value := range curMap {
			// ok will be false if the value is not in the map
			if val, ok := curMap[name]; ok {
				// only pay attention to values > 0, and that are not in the form post
				// the rest are just placeholders for days without blocks
				if val > 0 {
					if !form.Has(fmt.Sprintf("remove_block_%d_%s", x.ID, name)) {
						// delete the restriction by id
						err := m.DB.DeleteBlockByID(value)
						if err != nil {
							log.Println(err)
						}
					}
				}
			}
		}
	}

	// now handle new blocks
	for name, _ := range r.PostForm {
		if strings.HasPrefix(name, "add_block") {
			exploded := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(exploded[2])
			t, _ := time.Parse("2006-01-2", exploded[3])
			// insert a new block
			err := m.DB.InsertBlockForRoom(roomID, t)
			if err != nil {
				log.Println(err)
			}
		}
	}

	m.App.Session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}