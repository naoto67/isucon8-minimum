package main

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
)

type Event struct {
	ID       int64  `json:"id,omitempty"`
	Title    string `json:"title,omitempty"`
	PublicFg bool   `json:"public,omitempty"`
	ClosedFg bool   `json:"closed,omitempty"`
	Price    int64  `json:"price,omitempty"`

	Total   int                `json:"total"`
	Remains int                `json:"remains"`
	Sheets  map[string]*Sheets `json:"sheets,omitempty"`
}

func getEvents(all bool) ([]*Event, error) {
	cacheEvents, err := getEventsFromCache()
	if err != nil {
		return nil, err
	}
	sort.Slice(cacheEvents, func(i, j int) bool { return cacheEvents[i].ID < cacheEvents[j].ID })

	var events []*Event
	for i := range cacheEvents {
		if !all && !cacheEvents[i].PublicFg {
			continue
		}
		initializeEvent(&cacheEvents[i])

		ranks := []string{"S", "A", "B", "C"}
		for _, rank := range ranks {
			reservations, _ := getReservationsFromCache(cacheEvents[i].ID, rank)
			count := len(reservations)
			cacheEvents[i].Remains -= count
			cacheEvents[i].Sheets[rank].Remains -= count
		}

		events = append(events, &cacheEvents[i])
	}

	// rows, err = db.Query("SELECT * FROM reservations WHERE canceled_at IS NULL")
	// if err != nil {
	// 	return nil, err
	// }
	// defer rows.Close()
	// for rows.Next() {
	// 	var reservation Reservation
	// 	err = rows.Scan(&reservation.ID, &reservation.EventID, &reservation.SheetID, &reservation.UserID, &reservation.ReservedAt, &reservation.CanceledAt)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	event := getEventByID(events, reservation.EventID)
	// 	if event != nil {
	// 		err := assignReservation(event, reservation)
	// 		if err != nil {
	// 			return nil, err
	// 		}

	// 	}
	// }

	return events, nil
}

func getEventByID(events []*Event, id int64) *Event {
	for k := range events {
		if events[k].ID == id {
			return events[k]
		}
	}
	return nil
}

func getEvent(eventID, loginUserID int64) (*Event, error) {
	var event Event
	if err := db.QueryRow("SELECT * FROM events WHERE id = ?", eventID).Scan(&event.ID, &event.Title, &event.PublicFg, &event.ClosedFg, &event.Price); err != nil {
		return nil, err
	}
	event.Total = 1000
	event.Remains = 1000

	event.Sheets = map[string]*Sheets{
		"S": &Sheets{Total: 50, Remains: 50, Price: 5000 + event.Price},
		"A": &Sheets{Total: 150, Remains: 150, Price: 3000 + event.Price},
		"B": &Sheets{Total: 300, Remains: 300, Price: 1000 + event.Price},
		"C": &Sheets{Total: 500, Remains: 500, Price: 0 + event.Price},
	}

	// 1000席の初期化
	var i int64
	for i = 1; i <= 1000; i++ {
		sheet, ok := getSheetByID(i)
		if ok < 0 {
			return nil, errors.New("not found")
		}
		event.Sheets[sheet.Rank].Detail = append(event.Sheets[sheet.Rank].Detail, sheet)
	}

	// 予約席情報
	rows, err := db.Query("SELECT * FROM reservations WHERE event_id = ? AND canceled_at IS NULL", event.ID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var reservation Reservation
		err := rows.Scan(&reservation.ID, &reservation.EventID, &reservation.SheetID, &reservation.UserID, &reservation.ReservedAt, &reservation.CanceledAt)
		if err != nil {
			return nil, err
		}

		sheet, ok := getSheetByID(reservation.SheetID)
		if ok < 0 {
			return nil, errors.New("not found")
		}

		event.Sheets[sheet.Rank].Detail[sheet.Num-1].Mine = reservation.UserID == loginUserID
		event.Sheets[sheet.Rank].Detail[sheet.Num-1].Reserved = true
		event.Sheets[sheet.Rank].Detail[sheet.Num-1].ReservedAtUnix = reservation.ReservedAt.Unix()

		event.Sheets[sheet.Rank].Remains--
		event.Remains--
	}

	return &event, nil
}

func sanitizeEvent(e *Event) *Event {
	sanitized := *e
	sanitized.Price = 0
	sanitized.PublicFg = false
	sanitized.ClosedFg = false
	return &sanitized
}

func getEventWithoutDetail(event Event, loginUserID int64) (*Event, error) {
	event.Sheets = map[string]*Sheets{
		"S": &Sheets{Total: 50, Remains: 50, Price: 5000 + event.Price},
		"A": &Sheets{Total: 150, Remains: 150, Price: 3000 + event.Price},
		"B": &Sheets{Total: 300, Remains: 300, Price: 1000 + event.Price},
		"C": &Sheets{Total: 500, Remains: 500, Price: 0 + event.Price},
	}

	// 予約席情報
	rows, err := db.Query("SELECT * FROM reservations WHERE event_id = ? AND canceled_at IS NULL GROUP BY sheet_id HAVING reserved_at = MIN(reserved_at)", event.ID)
	if err != nil {
		return nil, err
	}

	event.Total = 1000
	event.Remains = 1000

	defer rows.Close()

	for rows.Next() {
		var reservation Reservation
		err := rows.Scan(&reservation.ID, &reservation.EventID, &reservation.SheetID, &reservation.UserID, &reservation.ReservedAt, &reservation.CanceledAt)
		if err != nil {
			return nil, err
		}

		sheet, ok := getSheetByID(reservation.SheetID)
		if ok < 0 {
			return nil, errors.New("not found")
		}

		event.Sheets[sheet.Rank].Remains--
		event.Remains--
	}

	return &event, nil
}

func assignReservation(event *Event, reservation Reservation) error {
	sheet, ok := getSheetByID(reservation.SheetID)
	if ok < 0 {
		return errors.New("not found")
	}
	event.Remains--
	event.Sheets[sheet.Rank].Remains--
	return nil
}

func initializeEvent(event *Event) {
	event.Total = 1000
	event.Remains = 1000
	event.Sheets = map[string]*Sheets{
		"S": &Sheets{Total: 50, Remains: 50, Price: 5000 + event.Price},
		"A": &Sheets{Total: 150, Remains: 150, Price: 3000 + event.Price},
		"B": &Sheets{Total: 300, Remains: 300, Price: 1000 + event.Price},
		"C": &Sheets{Total: 500, Remains: 500, Price: 0 + event.Price},
	}
}

func (e *Event) toJson() []byte {
	j := map[string]interface{}{
		"id":     e.ID,
		"title":  e.Title,
		"public": e.PublicFg,
		"closed": e.ClosedFg,
		"price":  e.Price,
	}
	data, _ := json.Marshal(j)
	return data
}

func initEvents() error {
	rows, err := db.Query("SELECT * FROM events")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var e Event
		err := rows.Scan(&e.ID, &e.Title, &e.PublicFg, &e.ClosedFg, &e.Price)
		if err != nil {
			return err
		}
		pushEventToCache(&e)
	}
	return nil
}

func pushEventToCache(event *Event) {
	key := EVENT_KEY
	data := event.toJson()
	setHashDataToCache(key, strconv.Itoa(int(event.ID)), data)
}

func getEventsFromCache() ([]Event, error) {
	data, err := getAllHashDataFromCache(EVENT_KEY)
	if err != nil {
		return nil, err
	}
	var events []Event
	err = json.Unmarshal(data, &events)

	return events, err
}
