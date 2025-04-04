package http

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"projeto_livros/internal/domain/models"

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

	// Check if a specific genre_id is requested
	genreID := r.URL.Query().Get("genre_id")
	if genreID != "" {
		// Return details for a specific genre
		var genreName, genreDescription string
		err := h.db.QueryRow(`
			SELECT name, description 
			FROM genres 
			WHERE id = $1`, genreID).Scan(&genreName, &genreDescription)
		if err == sql.ErrNoRows {
			http.Error(w, "Gênero não encontrado", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("Erro ao buscar gênero: %v", err)
			http.Error(w, "Erro ao buscar gênero", http.StatusInternalServerError)
			return
		}

		query := `
			SELECT l.id, l.name, l.quantity
			FROM livros l
			WHERE l.genre_id = $1
			ORDER BY l.name`
		rows, err := h.db.Query(query, genreID)
		if err != nil {
			log.Printf("Erro ao buscar livros: %v", err)
			http.Error(w, "Erro ao buscar livros", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var books []models.Book
		for rows.Next() {
			var book models.Book
			err := rows.Scan(&book.ID, &book.Name, &book.Quantity)
			if err != nil {
				log.Printf("Erro ao ler linha: %v", err)
				continue
			}
			book.GenreID = &genreID
			books = append(books, book)
		}

		genreWithBooks := models.GenreWithBooks{
			Name:       genreName,
			TotalBooks: len(books),
			Books:      books,
		}
		json.NewEncoder(w).Encode(genreWithBooks)
	} else {
		// Return all genres when no specific genre_id is provided
		query := `
			SELECT id, name, description 
			FROM genres 
			ORDER BY name`
		rows, err := h.db.Query(query)
		if err != nil {
			log.Printf("Erro ao buscar todos os gêneros: %v", err)
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
		SELECT l.id, l.name, l.quantity, l.genre_id, l.author 
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
		var authorNull sql.NullString

		if err := rows.Scan(&book.ID, &book.Name, &book.Quantity, &book.GenreID, &authorNull); err != nil {
			log.Printf("Erro ao ler livro: %v", err)
			continue
		}

		// Converter author de NullString para string
		if authorNull.Valid {
			book.Author = authorNull.String
		} else {
			book.Author = ""
		}

		// Garantir que title seja igual a name para compatibilidade
		book.Title = book.Name

		books = append(books, book)
	}
	json.NewEncoder(w).Encode(books)
}
