package dbrepository

import (
	"errors"
	"log"
	"time"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
)

func (pgRepo *testDBRepository) AllUsers() bool {
	return true
}

// Inserts a reservation into the database
func (pgRepo *testDBRepository) InsertReservation(reservation models.Reservation) (int, error) {
	// If room id is 2 then fail, otherwise pass
	if reservation.RoomID == 2 {
		return 0, errors.New("invalid room id")
	}

	return 1, nil
}

// Inserts a room restriction into the database
func (pgRepo *testDBRepository) InsertRoomRestriction(roomRestriction models.RoomRestriction) error {
	// If room id is equal to 1000 then fail, otherwise pass
	if roomRestriction.RoomID == 1000 {
		return errors.New("invalid room id")
	}

	return nil
}

// Queries for existing reservations on the given room and dates 
// Returns true if there are reservations for the given room and dates, otherwise it returns false
func (pgRepo *testDBRepository) SearchAvailabilityByDatesAndRoom(startDate time.Time, endDate time.Time, roomID int) (bool, error) {		
	// If the start date is after 2049-12-31, then return false
	// indicating that no rooms are available
	layout := "2006-01-02"
	limitDateStr := "2049-12-31"
	limitDate, err := time.Parse(layout, limitDateStr)
	if err != nil {
		log.Println(err)
	}

	if startDate.After(limitDate) {
		return false, nil
	}

	// This is our test to fail the query -- specify 2000-01-01 as start
	// A date in the past should not be valid
	testDateToFail, err := time.Parse(layout, "2000-01-01")
	if err != nil {
		log.Println(err)
	}

	if startDate == testDateToFail {
		return false, errors.New("invalid arrival date")
	}

	// If start date is valid, then just return true
	return true, nil
}

// Returns all rooms which are available during the given dates
func (pgRepo *testDBRepository) SearchAvailabilityForAllRooms(startDate time.Time, endDate time.Time) ([]models.Room, error) {
	var rooms []models.Room
	
	// If the start date is after 2049-12-31, then return empty slice
	// indicating that no rooms are available
	layout := "2006-01-02"
	limitDateStr := "2049-12-31"
	limitDate, err := time.Parse(layout, limitDateStr)
	if err != nil {
		log.Println(err)
	}

	if startDate.After(limitDate) {
		return rooms, nil
	}

	// This is our test to fail the query -- specify 2000-01-01 as start
	// A date in the past should not be valid
	testDateToFail, err := time.Parse(layout, "2000-01-01")
	if err != nil {
		log.Println(err)
	}

	if startDate == testDateToFail {
		return rooms, errors.New("invalid arrival date")
	}

	// If start date is valid, then put an entry into the `rooms` slice, indicating that some room is
	// available for search dates
	room := models.Room{
		ID: 1,
	}
	rooms = append(rooms, room)

	return rooms, nil
}

// Gets room by id
func (pgRepo *testDBRepository) GetRoomByID(id int) (models.Room, error) {
	var room models.Room

	// There are only 2 rooms with ID 1 and 2
	if id > 2 {
		return room, errors.New("can't find room with given id")
	}

	return room, nil
}