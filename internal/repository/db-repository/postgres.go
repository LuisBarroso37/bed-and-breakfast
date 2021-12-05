package dbrepository

import (
	"context"
	"errors"
	"time"

	"github.com/LuisBarroso37/bed-and-breakfast/internal/models"
	"golang.org/x/crypto/bcrypt"
)

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

// Gets user by id
func (pgRepo *postgresDBRepository) GetUserByID(id int) (models.User, error) {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	query := `SELECT id, first_name, last_name, email, password, access_level
		FROM users
		WHERE id = $1`

	var user models.User

	err := pgRepo.DB.QueryRowContext(
		ctx, 
		query, 
		id,
	).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.AccessLevel, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return user, err
	}

	return user, nil
}

// Updates a user
func (pgRepo *postgresDBRepository) UpdateUser(user models.User) error {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	query := `UPDATE users
		SET first_name = $1, last_name = $2, email = $3, access_level = $4, updated_at = $5`

	_, err := pgRepo.DB.ExecContext(
		ctx, 
		query, 
		user.FirstName, 
		user.LastName,
		user.Email,
		user.AccessLevel,
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}

// Authenticates a user
func (pgRepo *postgresDBRepository) Authenticate(email, password string) (int, string, error) {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	var id int // ID of authenticated user
	var hashedPassword string

	// Retrieve stored password corresponding to received email
	query := `SELECT id, password
		FROM users
		WHERE email = $1`

	err := pgRepo.DB.QueryRowContext(
		ctx, 
		query, 
		email,
	).Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	// Compare received password with stored password in the database
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}

// Gets a list of all reservations
func (pgRepo *postgresDBRepository) GetAllReservations() ([]models.Reservation, error) {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `SELECT r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, 
		r.end_date, r.room_id, r.created_at, r.updated_at, r.processed, rm.id, rm.room_name
		FROM reservations r
		LEFT JOIN rooms rm ON (r.room_id = rm.id)
		ORDER BY r.start_date ASC`

	rows, err := pgRepo.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}

	defer rows.Close()

	for rows.Next() {
		var reservation models.Reservation

		err := rows.Scan(
			&reservation.ID,
			&reservation.FirstName,
			&reservation.LastName,
			&reservation.Email,
			&reservation.Phone,
			&reservation.StartDate,
			&reservation.EndDate,
			&reservation.RoomID,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
			&reservation.Processed,
			&reservation.Room.ID,
			&reservation.Room.RoomName,
		)
		if err != nil {
			return reservations, nil
		}

		reservations = append(reservations, reservation)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// Gets a list of all new reservations (not yet processed)
func (pgRepo *postgresDBRepository) GetNewReservations() ([]models.Reservation, error) {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `SELECT r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, 
		r.end_date, r.room_id, r.created_at, r.updated_at, rm.id, rm.room_name
		FROM reservations r
		LEFT JOIN rooms rm ON (r.room_id = rm.id)
		WHERE processed = false
		ORDER BY r.start_date ASC`

	rows, err := pgRepo.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}

	defer rows.Close()

	for rows.Next() {
		var reservation models.Reservation

		err := rows.Scan(
			&reservation.ID,
			&reservation.FirstName,
			&reservation.LastName,
			&reservation.Email,
			&reservation.Phone,
			&reservation.StartDate,
			&reservation.EndDate,
			&reservation.RoomID,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
			&reservation.Room.ID,
			&reservation.Room.RoomName,
		)
		if err != nil {
			return reservations, nil
		}

		reservations = append(reservations, reservation)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

// Gets a reservation that matches given id
func (pgRepo *postgresDBRepository) GetReservationByID(id int) (models.Reservation, error) {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	var reservation models.Reservation

	query := `SELECT r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date,
		r.end_date, r.room_id, r.created_at, r.updated_at, r.processed, rm.id, rm.room_name
		FROM reservations r
		LEFT JOIN rooms rm ON (r.room_id = rm.id)
		WHERE r.id = $1`

	err := pgRepo.DB.QueryRowContext(
		ctx, 
		query,
		id,
	).Scan(
		&reservation.ID,
		&reservation.FirstName,
		&reservation.LastName,
		&reservation.Email,
		&reservation.Phone,
		&reservation.StartDate,
		&reservation.EndDate,
		&reservation.RoomID,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
		&reservation.Processed,
		&reservation.Room.ID,
		&reservation.Room.RoomName,
	)
	if err != nil {
		return reservation, err
	}

	return reservation, nil
}

// Updates a reservation
func (pgRepo *postgresDBRepository) UpdateReservation(reservation models.Reservation) error {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	query := `UPDATE reservations
		SET first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = $5
		WHERE id = $6`

	_, err := pgRepo.DB.ExecContext(
		ctx, 
		query, 
		reservation.FirstName, 
		reservation.LastName,
		reservation.Email,
		reservation.Phone,
		time.Now(),
		reservation.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// Deletes a reservation
func (pgRepo *postgresDBRepository) DeleteReservation(id int) error {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	query := `DELETE FROM reservations WHERE id = $1`

	_, err := pgRepo.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// Updates processed status of a reservation with given id
func (pgRepo *postgresDBRepository) UpdateProcessedForReservation(id int, processed bool) error {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	query := `UPDATE reservations
		SET processed = $1
		WHERE id = $2`

	_, err := pgRepo.DB.ExecContext(
		ctx, 
		query, 
		processed,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

// Gets all rooms
func (pgRepo *postgresDBRepository) GetAllRooms() ([]models.Room, error) {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	var rooms []models.Room

	query := `SELECT id, room_name, created_at, updated_at
		FROM rooms
		ORDER BY room_name`

	rows, err := pgRepo.DB.QueryContext(ctx, query)
	if err != nil {
		return rooms, err
	}

	defer rows.Close()

	for rows.Next() {
		var room models.Room

		err := rows.Scan(
			&room.ID,
			&room.RoomName,
			&room.CreatedAt,
			&room.UpdatedAt,
		)
		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// Gets restrictions for a given room by date range
func (pgRepo *postgresDBRepository) GetRestrictionsForRoomByDate(roomID int, startDate, endDate time.Time) ([]models.RoomRestriction, error) {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	var restrictions []models.RoomRestriction

	query := `SELECT id, COALESCE(reservation_id, 0), restriction_id, room_id, start_date, end_date
		FROM room_restrictions
		WHERE $1 < end_date and $2 >= start_date and room_id = $3`

	rows, err := pgRepo.DB.QueryContext(ctx, query, startDate, endDate, roomID)
	if err != nil {
		return restrictions, err
	}

	defer rows.Close()

	for rows.Next() {
		var restriction models.RoomRestriction

		err := rows.Scan(
			&restriction.ID,
			&restriction.ReservationID,
			&restriction.RestrictionID,
			&restriction.RoomID,
			&restriction.StartDate,
			&restriction.EndDate,
		)
		if err != nil {
			return restrictions, err
		}

		restrictions = append(restrictions, restriction)
	}

	if err = rows.Err(); err != nil {
		return restrictions, err
	}

	return restrictions, nil
}

// Inserts an owner block as room restriction for given room
func (pgRepo *postgresDBRepository) InsertBlockForRoom(id int, startDate time.Time) error {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	query := `INSERT INTO room_restrictions (start_date, end_date, room_id, restriction_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := pgRepo.DB.ExecContext(ctx, query, startDate, startDate.AddDate(0, 0, 1), id, 2, time.Now(), time.Now())
	if err != nil {
		return err
	}

	return nil
}

// Deletes an owner block from room restrictions
func (pgRepo *postgresDBRepository) DeleteBlockByID(id int) error {
	// Set timeout for this operation
	// Cancel operation if it takes more than 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	query := `DELETE FROM room_restrictions
		WHERE id = $1`

	_, err := pgRepo.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}