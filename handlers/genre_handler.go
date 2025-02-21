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

	genreID := r.URL.Query().Get("genre_id")

	var query string
	var rows *sql.Rows
	var err error

	if genreID != "" {
		query = `
			SELECT g.name, g.description, 
				   l.id, l.name, l.quantity
			FROM genres g
			LEFT JOIN livros l ON g.id = l.genre_id
			WHERE g.id = $1
			ORDER BY g.name`
		rows, err = h.db.Query(query, genreID)
	} else {
		query = `
			SELECT g.name, g.description, 
				   l.id, l.name, l.quantity
			FROM genres g
			LEFT JOIN livros l ON g.id = l.genre_id
			ORDER BY g.name`
		rows, err = h.db.Query(query)
	}

	if err != nil {
		log.Printf("Erro ao buscar gêneros: %v", err)
		http.Error(w, "Erro ao buscar gêneros", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	genresMap := make(map[string]*models.GenreWithBooks)

	for rows.Next() {
		var genreName, genreDescription string
		var bookID, bookName sql.NullString
		var bookQuantity sql.NullInt64

		err := rows.Scan(
			&genreName,
			&genreDescription,
			&bookID,
			&bookName,
			&bookQuantity,
		)
		if err != nil {
			log.Printf("Erro ao ler linha: %v", err)
			continue
		}

		if _, exists := genresMap[genreName]; !exists {
			genresMap[genreName] = &models.GenreWithBooks{
				Name:  genreName,
				Books: []models.Book{},
			}
		}

		if bookID.Valid {
			book := models.Book{
				ID:       bookID.String,
				Name:     bookName.String,
				Quantity: int(bookQuantity.Int64),
			}
			genresMap[genreName].Books = append(genresMap[genreName].Books, book)
		}
	}

	var genres []models.GenreWithBooks
	if genreID != "" {
		for _, g := range genresMap {
			genres = append(genres, *g)
			break
		}
	} else {
		for _, g := range genresMap {
			genres = append(genres, *g)
		}
	}

	if genres == nil {
		genres = []models.GenreWithBooks{}
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

func (h *GenreHandler) GetBooksByGenre(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	genreID := r.URL.Query().Get("genre_id")
	if genreID == "" {
		http.Error(w, "ID do gênero é obrigatório", http.StatusBadRequest)
		return
	}

	query := `
		SELECT l.id, l.name, l.quantity, l.genre_id 
		FROM livros l 
		WHERE l.genre_id = $1`

	rows, err := h.db.Query(query, genreID)
	if err != nil {
		log.Printf("Erro ao buscar livros por gênero: %v", err)
		http.Error(w, "Erro ao buscar livros", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.ID, &book.Name, &book.Quantity, &book.GenreID); err != nil {
			log.Printf("Erro ao ler livro: %v", err)
			continue
		}
		books = append(books, book)
	}

	json.NewEncoder(w).Encode(books)
}
