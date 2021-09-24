package dbrepository

import (
	"context"
	"time"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
)

func (pgRepo *postgresDBRepository) AllUsers() bool {
	return true
}

// Inserts a reservation into the database
func (pgRepo *postgresDBRepository) InsertReservation(reservation models.Reservation) (int, error) {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	query := `INSERT INTO reservations (first_name, last_name, email, phone, start_date,
						end_date, room_id, created_at, updated_at)
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
						RETURNING id`
					
	var reservationID int

	err := pgRepo.DB.QueryRowContext(
		ctx,
		query,
		reservation.FirstName,
		reservation.LastName,
		reservation.Email,
		reservation.Phone,
		reservation.StartDate,
		reservation.EndDate,
		reservation.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&reservationID)
	if err != nil {
		return 0, err
	}

	return reservationID, nil
}

// Inserts a room restriction into the database
func (pgRepo *postgresDBRepository) InsertRoomRestriction(roomRestriction models.RoomRestriction) error {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	query := `INSERT INTO room_restrictions (start_date, end_date, room_id, reservation_id, restriction_id, 
						created_at, updated_at)
						VALUES ($1, $2, $3, $4, $5, $6, $7)
						RETURNING id`

	_, err := pgRepo.DB.ExecContext(
		ctx,
		query,
		roomRestriction.StartDate,
		roomRestriction.EndDate,
		roomRestriction.RoomID,
		roomRestriction.ReservationID,
		roomRestriction.RestrictionID,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}

// Queries for existing reservations on the given room and dates 
// Returns true if there are reservations for the given room and dates, otherwise it returns false
func (pgRepo *postgresDBRepository) SearchAvailabilityByDatesAndRoom(startDate time.Time, endDate time.Time, roomID int) (bool, error) {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	query := `SELECT count(id)
						FROM room_restrictions
						WHERE room_id = $1
						AND $2 < end_date AND $3 > start_date`

	var numExistingReservations int

	err := pgRepo.DB.QueryRowContext(
		ctx,
		query,
		roomID,
		startDate,
		endDate,
	).Scan(&numExistingReservations)
	if err != nil {
		return false, err
	}

	if numExistingReservations == 0 {
		return true, nil
	}
	
	return false, nil
}

// Returns all rooms which are available during the given dates
func (pgRepo *postgresDBRepository) SearchAvailabilityForAllRooms(startDate time.Time, endDate time.Time) ([]models.Room, error) {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	query := `SELECT r.id, r.room_name
						FROM rooms r
						WHERE r.id NOT IN (SELECT room_id FROM room_restrictions rr WHERE $1 < end_date AND $2 > start_date)`

	var rooms []models.Room

	rows, err := pgRepo.DB.QueryContext(
		ctx,
		query,
		startDate,
		endDate,
	)
	if err != nil {
		return rooms, err
	}

	// Loop through each row returned from the query and add it to the `rooms` variable
	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.ID, &room.RoomName)
		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, room)
	}

	// Make sure no error occurred
	err = rows.Err()
	if err != nil {
		return rooms, err
	}
	
	return rooms, nil
}

// Gets room by id
func (pgRepo *postgresDBRepository) GetRoomByID(id int) (models.Room, error) {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	var room models.Room

	query := `SELECT id, room_name, created_at, updated_at
					FROM rooms
					WHERE id = $1`
				
	err := pgRepo.DB.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(&room.ID, &room.RoomName, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return room, err
	}

	return room, nil
}