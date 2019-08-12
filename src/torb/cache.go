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
