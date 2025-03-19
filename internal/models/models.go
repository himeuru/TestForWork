package models

import "time"

type Song struct {
	ID          int       `json:"id"`
	Group       string    `json:"group"`
	Song        string    `json:"song"`
	ReleaseDate time.Time `json:"release_date"`
	Link        string    `json:"link"`
	Lyrics      string    `json:"lyrics"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SongRequest struct {
	Group string `json:"group"`
	Song  string `json:"song"`
}

type SongDetail struct {
	ReleaseDate string `json:"release_date"`
	Lyrics      string `json:"lyrics"`
	Link        string `json:"link"`
}

type SongUpdateRequest struct {
	Group       *string `json:"group,omitempty"`
	Song        *string `json:"song,omitempty"`
	ReleaseDate *string `json:"release_date,omitempty"`
	Lyrics      *string `json:"lyrics,omitempty"`
	Link        *string `json:"link,omitempty"`
}
