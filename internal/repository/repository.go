package repository

import (
	"time"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
)

type DatabaseRepository interface {
	InsertReservation(reservation models.Reservation) (int, error)
	InsertRoomRestriction(roomRestriction models.RoomRestriction) error
	SearchAvailabilityByDatesAndRoom(startDate time.Time, endDate time.Time, roomID int) (bool, error)
	SearchAvailabilityForAllRooms(startDate time.Time, endDate time.Time) ([]models.Room, error)
	GetRoomByID(id int) (models.Room, error)
	GetUserByID(id int) (models.User, error)
	UpdateUser(user models.User) error
	Authenticate(email, password string) (int, string, error)
	GetAllReservations() ([]models.Reservation, error)
	GetNewReservations() ([]models.Reservation, error)
	GetReservationByID(id int) (models.Reservation, error)
	UpdateReservation(reservation models.Reservation) error
	DeleteReservation(id int) error
	UpdateProcessedForReservation(id int, processed bool) error
	GetAllRooms() ([]models.Room, error)
	GetRestrictionsForRoomByDate(roomID int, startDate, endDate time.Time) ([]models.RoomRestriction, error)
	InsertBlockForRoom(id int, startDate time.Time) error
	DeleteBlockByID(id int) error
}