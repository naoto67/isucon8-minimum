package main

import (
	"errors"
	"github.com/patrickmn/go-cache"
	"strconv"
	"time"
)

const (
	reservationCacheKeyPrefix = "RSVS-E-"
)

var (
	c = cache.New(5*time.Minute, 10*time.Minute)
)

func makeReservationCacheKey(eventID int64) string {
	return reservationCacheKeyPrefix + strconv.Itoa(int(eventID))
}

func setActiveReservationToCache(eventID int64, reservations []Reservation) {
	c.Set(makeReservationCacheKey(eventID), reservations, cache.DefaultExpiration)
}

func getActiveReservationFromCache(eventID int64) ([]Reservation, error) {
	key := makeReservationCacheKey(eventID)
	if reservations, found := c.Get(key); found {
		return reservations.([]Reservation), nil
	}
	return nil, errors.New("go-cache: key not found")
}

func appendActiveReservationToCache(eventID int64, reservation Reservation) error {
	reservations, err := getActiveReservationFromCache(eventID)
	if err != nil {
		newReservations := []Reservation{reservation}
		setActiveReservationToCache(eventID, newReservations)
	} else {
		reservations := append(reservations, reservation)
		setActiveReservationToCache(eventID, reservations)
	}
	return nil
}

func removeReservationFromCache(eventID, reservationID int64) error {
	old, err := getActiveReservationFromCache(eventID)
	if err != nil {
		return err
	}
	for i, r := range old {
		if r.ID == reservationID {
			newReservations := append(old[:i], old[i+1:]...)
			setActiveReservationToCache(eventID, newReservations)
			break
		}
	}
	return nil
}

func initActiveRerservations() error {
	rows, err := db.Query("SELECT * FROM reservations WHERE canceled_at IS NULL")
	if err != nil {
		return err
	}
	defer rows.Close()

	var activeReservationsMap map[int64][]Reservation

	var reservation Reservation
	for rows.Next() {
		err := rows.Scan(&reservation.ID, &reservation.EventID, &reservation.SheetID, &reservation.UserID, &reservation.ReservedAt, &reservation.CanceledAt)
		if err != nil {
			return err
		}
		activeReservationsMap[reservation.EventID] = append(activeReservationsMap[reservation.EventID], reservation)
	}
	for k, v := range activeReservationsMap {
		setActiveReservationToCache(k, v)
	}
	return nil
}
