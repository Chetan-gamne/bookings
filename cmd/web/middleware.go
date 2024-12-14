package main

import (
	"fmt"
	"net/http"

	"github.com/Chetan-gamne/bookings/internal/helpers"
	"github.com/justinas/nosurf"
)

func WriteToConsole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		fmt.Println("Hit the Page")
		next.ServeHTTP(w,r)
	})
}

// 
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

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		if !helpers.IsAuthenticated(r){
			session.Put(r.Context(),"error","Log in First!")
			http.Redirect(w,r,"/user/login",http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w,r)
	})
}