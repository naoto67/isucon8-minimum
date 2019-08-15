package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

var (
	redisHost = os.Getenv("REDIS_HOST")
	redisPort = os.Getenv("REDIS_PORT")

	Key                 = "KEY"
	ALL_RESERVATION_KEY = "ALL-RESERVAION-EVENT-ID-"
	EVENT_KEY           = "EVENT"
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

func getListDataFromCache(key string) ([]byte, error) {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return nil, err
	}

	strs, err := redis.Strings(conn.Do("LRANGE", key, 0, -1))
	if err != nil {
		return nil, err
	}
	str := strings.Join(strs[:], ",")
	str = "[" + str + "]"

	return []byte(str), err
}

// RPUSHは最後に追加
func pushListDataToCache(key string, data []byte) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return err
	}
	_, err = conn.Do("RPUSH", key, data)
	if err != nil {
		return err
	}
	return nil
}

// マッチするものを1つ削除
func removeListDataFromCache(key string, data []byte) error {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		return err
	}
	_, err = conn.Do("LREM", key, 1, data)
	if err != nil {
		return err
	}
	return nil
}

// =========================================================================

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
	for _, v := range reservations {
		data := (&v).toJson()
		pushListDataToCache(key, data)
	}
	return nil
}

func getReservationsFromCache(eventID int64, rank string) ([]Reservation, error) {
	var reservations []Reservation
	key := makeAllReservationsKey(eventID, rank)
	data, err := getListDataFromCache(key)
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
	data := (&reservation).toJson()
	key := makeAllReservationsKey(eventID, sheet.Rank)
	pushListDataToCache(key, data)
	return nil
}

func removeReservationFromCache(eventID int64, reservation Reservation) error {
	sheet, ok := getSheetByID(reservation.SheetID)
	if ok < 0 {
		return errors.New("not found")
	}
	data := (&reservation).toJson()
	key := makeAllReservationsKey(eventID, sheet.Rank)
	removeListDataFromCache(key, data)
	return nil
}

func initEvents() {
	rows, err := db.Query("SELECT * FROM events")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var e *Event
		err := rows.Scan(e.ID, e.Title, e.PublicFg, e.ClosedFg, e.Price)
		if err != nil {
			panic(err)
		}
		pushEventToCache(e)
	}
}

func pushEventToCache(event *Event) {
	key := EVENT_KEY
	data := event.toJson()
	pushListDataToCache(key, data)
}
