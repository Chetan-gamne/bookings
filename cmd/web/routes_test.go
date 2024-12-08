package main

import (
	"testing"

	"github.com/Chetan-gamne/bookings/internal/config"
	"github.com/go-chi/chi/v5"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)

	switch mux.(type) {
	case *chi.Mux:
		// do nothing test passed
	default:
		t.Error("type if not *chi.Mux")
		
	}
}