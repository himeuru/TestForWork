package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testForWork/internal/models"
	"time"
)

type Service struct {
	db       *sql.DB
	musicAPI string
}

const dateFormat = "2006-01-02"

func NewService(db *sql.DB, musicAPI string) *Service {
	return &Service{
		db:       db,
		musicAPI: musicAPI,
	}
}

func (service *Service) CreateSong(group, song string) (*models.Song, error) {
	if group == "" || song == "" {
		return nil, fmt.Errorf("group and song names cannot be empty")
	}

	log.Printf("Starting CreateSong for %s - %s", group, song)
	details := service.generateLocalDetails(group, song)

	// для подключенного API ->
	/*details, err := service.fetchSongDetails(group, song)
	if err != nil {
		log.Printf("Error fetching details: %v", err)
		return nil, fmt.Errorf("failed to fetch song details: %w", err)
	}*/

	log.Printf("Received details: %+v", details)

	releaseDate, err := time.Parse(dateFormat, details.ReleaseDate)
	if err != nil {
		log.Printf("Date parsing error: %v | Input: %s", err, details.ReleaseDate)
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	query := `INSERT INTO songs 
    			("group_name", "song_name", "release_date", "lyrics", "link") 
			  VALUES 
			    ($1, $2, $3, $4, $5) 
			  RETURNING id, created_at, updated_at`

	var newSong models.Song
	err = service.db.QueryRowContext(
		context.Background(), query, group, song, releaseDate, details.Lyrics, details.Link,
	).Scan(
		&newSong.ID, &newSong.CreatedAt, &newSong.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}

	newSong.Group = group
	newSong.Song = song
	newSong.ReleaseDate = releaseDate
	newSong.Lyrics = details.Lyrics
	newSong.Link = details.Link

	log.Printf("Created new song : %d", newSong)
	return &newSong, nil
}

func (service *Service) GetSongs(group, song string, page int, limit int) ([]models.Song, error) {
	offset, query := (page-1)*limit, `
		SELECT id, group_name, song_name, release_date, lyrics, link, created_at, updated_at 
		FROM songs
		WHERE ($1 = '' OR group_name = $1) AND ($2 = '' OR song_name = $2) LIMIT $3 OFFSET $4`

	rows, err := service.db.QueryContext(context.Background(), query, group, song, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer rows.Close()

	var songs []models.Song
	for rows.Next() {
		var song models.Song
		err := rows.Scan(
			&song.ID,
			&song.Group,
			&song.Song,
			&song.ReleaseDate,
			&song.Lyrics,
			&song.Link,
			&song.CreatedAt,
			&song.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		songs = append(songs, song)
	}
	return songs, nil
}

func (service *Service) GetLyrics(id, page, limit int) ([]string, error) {
	query := `SELECT lyrics FROM songs WHERE id=$1`
	var lyrics string
	err := service.db.QueryRowContext(context.Background(), query, id).Scan(&lyrics)
	if err != nil {
		return nil, err
	}

	verses := strings.Split(lyrics, "\n\n")
	start := (page - 1) * limit
	end := start + limit

	if start > len(verses) {
		return []string{}, nil
	}
	if end > len(verses) {
		end = len(verses)
	}

	return verses[start:end], nil
}

func (service *Service) UpdateSong(id int, req models.SongUpdateRequest) (*models.Song, error) {
	currentSong, err := service.GetSongByID(id)
	if err != nil {
		return nil, err
	}

	var updates []string
	var params []interface{}
	counter := 1

	if req.Group != nil {
		updates = append(updates, fmt.Sprintf("group_name = $%d", counter))
		params = append(params, *req.Group)
		counter++
	}
	if req.Song != nil {
		updates = append(updates, fmt.Sprintf("song_name = $%d", counter))
		params = append(params, *req.Song)
		counter++
	}
	if req.ReleaseDate != nil {
		rd, err := time.Parse(dateFormat, *req.ReleaseDate)
		if err != nil {
			return nil, fmt.Errorf("invalid release date format: %w", err)
		}
		updates = append(updates, fmt.Sprintf("release_date = $%d", counter))
		params = append(params, rd)
		counter++
	}
	if req.Lyrics != nil {
		updates = append(updates, fmt.Sprintf("lyrics = $%d", counter))
		params = append(params, *req.Lyrics)
		counter++
	}
	if req.Link != nil {
		updates = append(updates, fmt.Sprintf("link = $%d", counter))
		params = append(params, *req.Link)
		counter++
	}

	if len(updates) == 0 {
		return currentSong, nil
	}

	query := fmt.Sprintf(
		"UPDATE songs SET %s WHERE id = $%d RETURNING *",
		strings.Join(updates, ", "),
		counter,
	)
	params = append(params, id)

	// Выполняем запрос
	var updatedSong models.Song
	err = service.db.QueryRow(query, params...).Scan(
		&updatedSong.ID,
		&updatedSong.Group,
		&updatedSong.Song,
		&updatedSong.ReleaseDate,
		&updatedSong.Lyrics,
		&updatedSong.Link,
		&updatedSong.CreatedAt,
		&updatedSong.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("database update failed: %w", err)
	}

	return &updatedSong, nil
}

func (service *Service) DeleteSong(id int) error {
	query := `DELETE FROM songs WHERE id = $1`
	response, err := service.db.ExecContext(context.Background(), query, id)
	if err != nil {
		return fmt.Errorf("Database delete failed: %w", err)
	}

	rowsAffected, _ := response.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	log.Printf("Deleted song with id %d", id)
	return nil
}

func (service *Service) generateLocalDetails(group, song string) *models.SongDetail {
	return &models.SongDetail{
		ReleaseDate: time.Now().Format("2006-01-02"),
		Lyrics:      "Lyrics placeholder for " + song,
		Link:        fmt.Sprintf("https://example.com/%s/%s", group, song),
	}
}

func (service *Service) fetchSongDetails(group, song string) (*models.SongDetail, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	requestURL := fmt.Sprintf("%s/info?group=%s&song=%s",
		service.musicAPI,
		url.QueryEscape(group),
		url.QueryEscape(song))
	log.Printf("Fetching song details for %s", requestURL)

	response, err := client.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("API query failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %s", response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response: %w", err)
	}

	var detail models.SongDetail
	if err := json.Unmarshal(body, &detail); err != nil {
		return nil, fmt.Errorf("Failed to parse response: %w", err)
	}

	return &detail, nil
}

func (service *Service) GetSongByID(id int) (*models.Song, error) {
	query := `SELECT id, group_name, song_name, release_date, lyrics, link, created_at, updated_at FROM songs WHERE id = $1`

	var song models.Song
	err := service.db.QueryRow(query, id).Scan(
		&song.ID,
		&song.Group,
		&song.Song,
		&song.ReleaseDate,
		&song.Lyrics,
		&song.Link,
		&song.CreatedAt,
		&song.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("song not found: %w", err)
	}
	return &song, nil
}
