package models

import "time"

type URLRequest struct {
	Url    string  `json:"url"`
	Custom *string `json:"custom"`
}

type URL struct {
	ID        int64     `json:"-"`
	Original  string    `json:"original"`
	CreatedAt time.Time `json:"created_at"`
	Short     string    `sql:"-",json:"short"`
}

type URLStats struct {
	ID        int64     `json:"-"`
	UrlID     int64     `json:"-"`
	Date      time.Time `json:"date"`
	Hits      int       `json:"hits"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Stats struct {
	Original  string     `json:"original"`
	Short     string     `json:"short"`
	Total     int        `json:"total"`
	Histogram []URLStats `json:"histogram"`
	CreatedAt time.Time  `json:"created_at"`
}
