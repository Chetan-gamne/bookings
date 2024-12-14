package helpers

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/Chetan-gamne/bookings/internal/config"
)

var app *config.AppConfig

func NewHelpers(a *config.AppConfig) {
	app = a
}

func ClientError(w http.ResponseWriter, status int) {
	app.InfoLog.Println("Client Error With Status of",status)
	http.Error(w,http.StatusText(status),status);
}

func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s",err.Error(),debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(w,http.StatusText(http.StatusInternalServerError),http.StatusInternalServerError)
}

func IsAuthenticated(r *http.Request) bool {
	exists := app.Session.Exists(r.Context(),"user_id")

	return exists
}