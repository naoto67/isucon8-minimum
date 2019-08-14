package main

import "time"

type Reservation struct {
	ID         int64      `json:"id"`
	EventID    int64      `json:"event_id"`
	SheetID    int64      `json:"sheet_id"`
	UserID     int64      `json:"user_id"`
	ReservedAt *time.Time `json:"reserved_at"`
	CanceledAt *time.Time `json:"-"`

	Event          *Event `json:"event,omitempty"`
	SheetRank      string `json:"sheet_rank,omitempty"`
	SheetNum       int64  `json:"sheet_num,omitempty"`
	Price          int64  `json:"price,omitempty"`
	ReservedAtUnix int64  `json:"reserved_at,omitempty"`
	CanceledAtUnix int64  `json:"canceled_at,omitempty"`
}
