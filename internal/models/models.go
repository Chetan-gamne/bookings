package models

import "time"

// Users in user model
type User struct {
	ID int 
	FirstName string 
	LastName string
	Email string
	Password string 
	AccessLevel int 
	CreatedAt time.Time
	UpdatedAt time.Time 
}

// Rooms is the room model
type Room struct {
	ID int 
	RoomName string 
	CreatedAt time.Time 
	UpdatedAt time.Time 
}

// Restrictions is Restriction Model
type Restriction struct {
	ID int 
	RestrictionName string 
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Reservations is Reservation Model
type Reservation struct {
	ID int 
	FirstName string
	LastName string
	Email string
	Phone string 
	StartDate time.Time
	EndDate time.Time
	RoomID int
	CreatedAt time.Time
	UpdatedAt time.Time
	Room Room
}

// RoomRestrictions is room restriction Model
type RoomRestriction struct {
	ID	int 
	StartDate time.Time
	EndDate time.Time
	RoomID int
	ReservationID int
	RestrictionID int
	CreatedAt time.Time
	UpdatedAt time.Time
	Room Room
	Reservation Reservation
	Restriction Restriction
}

