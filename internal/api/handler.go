package api

import (
	"database/sql"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"strconv"
	"testForWork/internal/models"
	"testForWork/internal/service"
)

type Handler struct {
	service  *service.Service
	musicAPI string
}

func NewHandler(db *sql.DB, musicAPI string) *Handler {
	return &Handler{
		service:  service.NewService(db, musicAPI),
		musicAPI: musicAPI,
	}
}

// @Summary Get songs
// @Description Get songs with filtering and pagination
// @Tags songs
// @Accept json
// @Produce json
// @Param group query string false "Group filter"
// @Param song query string false "Song filter"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {array} models.Song
// @Router /songs [get]
func (handler *Handler) getSongs(writer http.ResponseWriter, router *http.Request) {
	group := router.URL.Query().Get("group")
	songName := router.URL.Query().Get("song_name")

	page, _ := strconv.Atoi(router.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(router.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	songs, err := handler.service.GetSongs(group, songName, page, limit)
	if err != nil {
		log.Printf("Error getting songs: %s\n", err)
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(songs)
}

// @Summary Get lyrics
// @Description Get paginated song lyrics
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {array} string
// @Router /songs/{id}/lyrics [get]
func (handler *Handler) getLyrics(writer http.ResponseWriter, router *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(router, "id"))

	page, _ := strconv.Atoi(router.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(router.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	lyrics, err := handler.service.GetLyrics(id, page, limit)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(writer, "Song not found", http.StatusNotFound)
			return
		}

		log.Printf("Error getting lyrics: %s\n", err)
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(lyrics)
}

// @Summary Add song
// @Description Add new song
// @Tags songs
// @Accept json
// @Produce json
// @Param song body models.SongRequest true "Song data"
// @Success 201 {object} models.Song
// @Router /songs [post]
func (handler *Handler) addSong(writer http.ResponseWriter, router *http.Request) {
	var request models.SongRequest
	if err := json.NewDecoder(router.Body).Decode(&request); err != nil {
		log.Printf("Error decoding request: %s\n", err)
		http.Error(writer, "Bad Request", http.StatusBadRequest)
		return
	}

	newSong, err := handler.service.CreateSong(request.Group, request.Song)
	log.Printf("successfully added song\n")
	if err != nil {
		log.Printf("Error creating song: %s\n", err)
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(newSong)
}

// @Summary Update song
// @Description Update song details
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Param song body models.Song true "Song data"
// @Success 200 {object} models.Song
// @Router /songs/{id} [put]
func (handler *Handler) updateSong(writer http.ResponseWriter, router *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(router, "id"))

	var request models.SongUpdateRequest
	if err := json.NewDecoder(router.Body).Decode(&request); err != nil {
		log.Printf("Error decoding request: %s\n", err)
		http.Error(writer, "Bad Request", http.StatusBadRequest)
		return
	}

	updatedSong, err := handler.service.UpdateSong(id, request)
	if err != nil {
		log.Printf("Error updating song: %s\n", err)
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(updatedSong)
}

// @Summary Delete song
// @Description Delete a song
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Success 204
// @Router /songs/{id} [delete]
func (handler *Handler) deleteSong(writer http.ResponseWriter, router *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(router, "id"))

	if err := handler.service.DeleteSong(id); nil != err {
		log.Printf("Error deleting song: %s\n", err)
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
}

func (handler *Handler) Routes() chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	router.Route("/songs", func(r chi.Router) {
		r.Get("/", handler.getSongs)
		r.Post("/", handler.addSong)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/lyrics", handler.getLyrics)
			r.Put("/", handler.updateSong)
			r.Delete("/", handler.deleteSong)
		})
	})

	return router
}
