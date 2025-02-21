package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"projeto_livros/models"

	"github.com/segmentio/ksuid"
)

type GenreHandler struct {
	db *sql.DB
}

func NewGenreHandler(db *sql.DB) *GenreHandler {
	return &GenreHandler{db: db}
}

func (h *GenreHandler) GetAllGenres(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := h.db.Query("SELECT id, name, description FROM genres")
	if err != nil {
		log.Printf("Erro ao buscar gêneros: %v", err)
		http.Error(w, "Erro ao buscar gêneros", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var genres []models.Genre
	for rows.Next() {
		var genre models.Genre
		if err := rows.Scan(&genre.ID, &genre.Name, &genre.Description); err != nil {
			log.Printf("Erro ao ler gênero: %v", err)
			continue
		}
		genres = append(genres, genre)
	}

	json.NewEncoder(w).Encode(genres)
}

func (h *GenreHandler) CreateGenre(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var genre models.Genre
	if err := json.NewDecoder(r.Body).Decode(&genre); err != nil {
		http.Error(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}

	genre.ID = ksuid.New().String()

	_, err := h.db.Exec("INSERT INTO genres (id, name, description) VALUES ($1, $2, $3)",
		genre.ID, genre.Name, genre.Description)
	if err != nil {
		log.Printf("Erro ao criar gênero: %v", err)
		http.Error(w, "Erro ao criar gênero", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(genre)
}
