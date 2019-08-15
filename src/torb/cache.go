package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/gomodule/redigo/redis"
)

var (
	redisHost = os.Getenv("REDIS_HOST")
	redisPort = os.Getenv("REDIS_PORT")

	Key                 = "KEY"
	ALL_RESERVATION_KEY = "ALL-RESERVAION-EVENT-ID-"
)

func flushALL() error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return err
	}
	conn.Do("FLUSHALL")
	return nil
}

func getDataFromCache(key string) ([]byte, error) {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return nil, err
	}

	data, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func setDataToCache(key string, data []byte) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return err
	}
	_, err = conn.Do("SET", key, data)
	if err != nil {
		return err
	}
	return nil
}

func makeKey(id int64) string {
	ID := strconv.Itoa(int(id))
	return Key + ID
}

// イベントごとの全ての予約（canceled_atがnullのもの）
func makeAllReservationsKey(eventID int64, rank string) string {
	ID := strconv.Itoa(int(eventID))
	return ALL_RESERVATION_KEY + ID + "-" + rank
}

func initAllReservations() {
	rows, err := db.Query("SELECT * FROM reservations WHERE canceled_at IS NULL")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	event_rank_reservations := make(map[int64]map[string][]Reservation)
	for rows.Next() {
		var r Reservation
		err := rows.Scan(&r.ID, &r.EventID, &r.SheetID, &r.UserID, &r.ReservedAt, &r.CanceledAt)
		if err != nil {
			fmt.Println(err)
		}
		sheet, ok := getSheetByID(r.SheetID)
		if ok < 0 {
			continue
		}
		if event_rank_reservations[r.EventID] == nil {
			event_rank_reservations[r.EventID] = make(map[string][]Reservation)
		}
		event_rank_reservations[r.EventID][sheet.Rank] = append(event_rank_reservations[r.EventID][sheet.Rank], r)
	}
	for key, value := range event_rank_reservations {
		for k, v := range value {
			err := setReservationsToCache(key, k, v)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func setReservationsToCache(eventID int64, rank string, reservations []Reservation) error {
	key := makeAllReservationsKey(eventID, rank)
	data, err := json.Marshal(reservations)
	if err != nil {
		return err
	}
	err = setDataToCache(key, data)
	return err
}

func getReservationsFromCache(eventID int64, rank string) ([]Reservation, error) {
	var reservations []Reservation
	key := makeAllReservationsKey(eventID, rank)
	data, err := getDataFromCache(key)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &reservations)
	if err != nil {
		return nil, err
	}
	return reservations, nil
}

func appendReservationToCache(eventID int64, reservation Reservation) error {
	sheet, ok := getSheetByID(reservation.SheetID)
	if ok < 0 {
		return errors.New("not found")
	}
	reservations, err := getReservationsFromCache(eventID, sheet.Rank)
	if err != nil {
		if err == redis.ErrNil {
			setReservationsToCache(eventID, sheet.Rank, []Reservation{reservation})
			return nil
		} else {
			return err
		}
	}
	reservations = append(reservations, reservation)
	setReservationsToCache(eventID, sheet.Rank, reservations)
	return nil
}

func removeReservationFromCache(eventID, reservationID int64, rank string) error {
	reservations, err := getReservationsFromCache(eventID, rank)
	if err != nil {
		return err
	}
	for i, v := range reservations {
		if v.ID == reservationID {
			newReservations := append(reservations[:i], reservations[i+1:]...)
			setReservationsToCache(eventID, rank, newReservations)
			break
		}
	}
	return nil
}
