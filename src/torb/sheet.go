package main

import (
	"time"
)

type Sheets struct {
	Total   int      `json:"total"`
	Remains int      `json:"remains"`
	Detail  []*Sheet `json:"detail,omitempty"`
	Price   int64    `json:"price"`
}

type Sheet struct {
	ID    int64  `json:"-"`
	Rank  string `json:"-"`
	Num   int64  `json:"num"`
	Price int64  `json:"-"`

	Mine           bool       `json:"mine,omitempty"`
	Reserved       bool       `json:"reserved,omitempty"`
	ReservedAt     *time.Time `json:"-"`
	ReservedAtUnix int64      `json:"reserved_at,omitempty"`
}

func validateRank(rank string) bool {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM sheets WHERE `rank` = ?", rank).Scan(&count)
	return count > 0
}

func getSheetByNumAndRank(num int64, rank string) (*Sheet, int64) {
	switch rank {
	case "S":
		return &Sheet{
			ID:    num,
			Num:   num,
			Price: 5000,
			Rank:  rank,
		}, 1
	case "A":
		return &Sheet{
			ID:    num + 50,
			Num:   num,
			Price: 3000,
			Rank:  rank,
		}, 1
	case "B":
		return &Sheet{
			ID:    num + 200,
			Num:   num,
			Price: 1000,
			Rank:  rank,
		}, 1

	case "C":
		return &Sheet{
			ID:    num + 500,
			Num:   num,
			Price: 0,
			Rank:  rank,
		}, 1
	}
	return nil, -1
}

func getSheetByID(id int64) (*Sheet, int64) {
	if 1 <= id && id <= 50 {
		return &Sheet{
			ID:    id,
			Num:   id,
			Price: 5000,
			Rank:  "S",
		}, 1
	} else if 51 <= id && id <= 200 {
		return &Sheet{
			ID:    id,
			Num:   id - 50,
			Price: 3000,
			Rank:  "A",
		}, 1
	} else if 201 <= id && id <= 500 {
		return &Sheet{
			ID:    id,
			Num:   id - 200,
			Price: 1000,
			Rank:  "B",
		}, 1
	} else if 501 <= id && id <= 1000 {
		return &Sheet{
			ID:    id,
			Num:   id - 500,
			Price: 0,
			Rank:  "C",
		}, 1
	}
	return nil, -1
}
