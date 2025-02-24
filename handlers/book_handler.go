package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"projeto_livros/models"
	"strconv"

	"github.com/segmentio/ksuid"
)

type BookHandler struct {
	db *sql.DB
}
type IDRequest struct {
	ID string `json:"id"`
}

func NewBookHandler(db *sql.DB) *BookHandler {
	return &BookHandler{db: db}
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		log.Printf("Erro ao decodificar JSON: %v", err)
		http.Error(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}

	// Validações
	if book.Name == "" || book.Quantity < 0 {
		http.Error(w, "Nome vazio ou quantidade negativa não permitidos", http.StatusBadRequest)
		return
	}

	// Verificar se o gênero existe
	if book.GenreID != "" {
		var count int
		err := h.db.QueryRow("SELECT COUNT(*) FROM genres WHERE id = $1", book.GenreID).Scan(&count)
		if err != nil || count == 0 {
			http.Error(w, "Gênero não encontrado", http.StatusBadRequest)
			return
		}
	}

	book.ID = ksuid.New().String()

	query := `INSERT INTO livros (id, name, quantity, genre_id) VALUES ($1, $2, $3, $4) RETURNING id`
	var returnedID string
	if err := h.db.QueryRow(query, book.ID, book.Name, book.Quantity, book.GenreID).Scan(&returnedID); err != nil {
		log.Printf("Erro ao inserir livro: %v", err)
		http.Error(w, "Erro ao criar livro", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) GetAllBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parâmetros de paginação com valores padrão
	page := 1
	perPage := 10

	// Obter parâmetros da query
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	offset := (page - 1) * perPage
	genreID := r.URL.Query().Get("genre_id")

	// Contar total de registros
	var total int
	countQuery := "SELECT COUNT(*) FROM livros"
	if genreID != "" {
		countQuery += " WHERE genre_id = $1"
		if err := h.db.QueryRow(countQuery, genreID).Scan(&total); err != nil {
			log.Printf("Erro ao contar livros: %v", err)
			http.Error(w, "Erro ao buscar livros", http.StatusInternalServerError)
			return
		}
	} else {
		if err := h.db.QueryRow(countQuery).Scan(&total); err != nil {
			log.Printf("Erro ao contar livros: %v", err)
			http.Error(w, "Erro ao buscar livros", http.StatusInternalServerError)
			return
		}
	}

	// Query base
	baseQuery := `
		SELECT l.id, l.name, l.quantity, l.genre_id 
		FROM livros l`

	if genreID != "" {
		baseQuery += " WHERE l.genre_id = $1"
	}

	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d",
		map[bool]int{true: 2, false: 1}[genreID != ""],
		map[bool]int{true: 3, false: 2}[genreID != ""])

	var rows *sql.Rows
	var err error
	if genreID != "" {
		rows, err = h.db.Query(baseQuery, genreID, perPage, offset)
	} else {
		rows, err = h.db.Query(baseQuery, perPage, offset)
	}

	if err != nil {
		log.Printf("Erro ao buscar livros: %v", err)
		http.Error(w, "Erro ao buscar livros", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.ID, &book.Name, &book.Quantity, &book.GenreID); err != nil {
			log.Printf("Erro ao ler dados do livro: %v", err)
			continue
		}
		books = append(books, book)
	}

	totalPages := (total + perPage - 1) / perPage

	response := models.PaginationResponse{
		Data: books,
	}
	response.Pagination.CurrentPage = page
	response.Pagination.PerPage = perPage
	response.Pagination.TotalItems = total
	response.Pagination.TotalPages = totalPages
	response.Pagination.HasPrevious = page > 1
	response.Pagination.HasNext = page < totalPages

	json.NewEncoder(w).Encode(response)
}

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	var req IDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Erro ao decodificar JSON: %v", err)
		http.Error(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		http.Error(w, "ID não fornecido", http.StatusBadRequest)
		return
	}

	var book models.Book
	err := h.db.QueryRow("SELECT id, name, quantity FROM livros WHERE id = $1", req.ID).
		Scan(&book.ID, &book.Name, &book.Quantity)
	if err == sql.ErrNoRows {
		http.Error(w, "Livro não encontrado", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Erro ao buscar livro: %v", err)
		http.Error(w, "Erro ao buscar livro", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	var req IDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Erro ao decodificar JSON: %v", err)
		http.Error(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		http.Error(w, "ID não fornecido", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec("DELETE FROM livros WHERE id = $1", req.ID)
	if err != nil {
		log.Printf("Erro ao deletar livro: %v", err)
		http.Error(w, "Erro ao deletar livro", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Livro não encontrado", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	// Decodificar o livro atualizado do body
	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		log.Printf("Erro ao decodificar JSON: %v", err)
		http.Error(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}

	if book.ID == "" {
		http.Error(w, "ID não fornecido", http.StatusBadRequest)
		return
	}
	if book.Name == "" || book.Quantity < 0 {
		http.Error(w, "Nome vazio ou quantidade negativa não permitidos", http.StatusBadRequest)
		return
	}

	// Atualizar o livro no banco de dados
	query := `UPDATE livros SET name = $1, quantity = $2 WHERE id = $3`
	result, err := h.db.Exec(query, book.Name, book.Quantity, book.ID)
	if err != nil {
		log.Printf("Erro ao atualizar livro: %v", err)
		http.Error(w, "Erro ao atualizar livro", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Livro não encontrado", http.StatusNotFound)
		return
	}

	// Retornar o livro atualizado
	response := map[string]string{
		"message": fmt.Sprintf("Livro '%s' atualizado com sucesso", book.Name),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *BookHandler) CreateAllBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()

	var books []models.Book
	if err := json.NewDecoder(r.Body).Decode(&books); err != nil {
		log.Printf("Erro ao decodificar JSON: %v", err)
		http.Error(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}

	var createdBooks []models.Book

	for _, book := range books {
		if book.Name == "" || book.Quantity < 0 {
			http.Error(w, "Nome vazio ou quantidade negativa não permitidos", http.StatusBadRequest)
			return
		}

		newID := ksuid.New().String()

		query := `INSERT INTO livros (id, name, quantity) VALUES ($1, $2, $3) RETURNING id`
		var returnedID string

		if err := h.db.QueryRow(query, newID, book.Name, book.Quantity).Scan(&returnedID); err != nil {
			log.Printf("Erro ao inserir livro: %v", err)
			http.Error(w, "Erro ao criar livro", http.StatusInternalServerError)
			return
		}

		book.ID = returnedID
		createdBooks = append(createdBooks, book)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdBooks)
}
