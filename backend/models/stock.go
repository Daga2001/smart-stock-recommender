package models

import "time"

type Stock struct {
	ID         int       `json:"id" db:"id"`
	Ticker     string    `json:"ticker" db:"ticker"`
	TargetFrom string    `json:"target_from" db:"target_from"`
	TargetTo   string    `json:"target_to" db:"target_to"`
	Company    string    `json:"company" db:"company"`
	Action     string    `json:"action" db:"action"`
	Brokerage  string    `json:"brokerage" db:"brokerage"`
	RatingFrom string    `json:"rating_from" db:"rating_from"`
	RatingTo   string    `json:"rating_to" db:"rating_to"`
	Time       time.Time `json:"time" db:"time"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type ApiResponse struct {
	Items    []Stock `json:"items"`
	NextPage string  `json:"next_page"`
}

type PageRequest struct {
	Page int `json:"page" binding:"required"`
}