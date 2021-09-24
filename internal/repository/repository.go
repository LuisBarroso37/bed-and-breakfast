package repository

import (
	"time"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
)

type DatabaseRepository interface {
	AllUsers() bool
	InsertReservation(reservation models.Reservation) (int, error)
	InsertRoomRestriction(roomRestriction models.RoomRestriction) error
	SearchAvailabilityByDatesAndRoom(startDate time.Time, endDate time.Time, roomID int) (bool, error)
	SearchAvailabilityForAllRooms(startDate time.Time, endDate time.Time) ([]models.Room, error)
	GetRoomByID(id int) (models.Room, error)
}