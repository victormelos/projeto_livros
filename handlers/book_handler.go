package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"projeto_livros/models"
	"strconv"
	"github.com/go-chi/chi/v5"
	"github.com/segmentio/ksuid"
)

type BookHandler struct {
	db *sql.DB
}

type IDRequest struct {
	ID string `json:"id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// Função auxiliar para enviar respostas de erro em formato JSON
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := ErrorResponse{
		Error: message,
		Code:  statusCode,
	}
	json.NewEncoder(w).Encode(err)
}

func NewBookHandler(db *sql.DB) *BookHandler {
	return &BookHandler{db: db}
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		log.Printf("Erro ao decodificar JSON: %v", err)
		sendErrorResponse(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}
	
	// Check if title is provided but name is not
	if book.Name == "" && book.Title != "" {
		book.Name = book.Title
	}
	
	if book.Name == "" || book.Quantity < 0 {
		sendErrorResponse(w, "Nome vazio ou quantidade negativa não permitidos", http.StatusBadRequest)
		return
	}
	if book.GenreID != nil {
		var count int
		err := h.db.QueryRow("SELECT COUNT(*) FROM genres WHERE id = $1", book.GenreID).Scan(&count)
		if err != nil || count == 0 {
			sendErrorResponse(w, "Gênero não encontrado", http.StatusBadRequest)
			return
		}
	}
	book.ID = ksuid.New().String()
	query := `INSERT INTO livros (id, name, quantity, genre_id) VALUES ($1, $2, $3, $4) RETURNING id`
	var returnedID string
	if err := h.db.QueryRow(query, book.ID, book.Name, book.Quantity, book.GenreID).Scan(&returnedID); err != nil {
		log.Printf("Erro ao inserir livro: %v", err)
		sendErrorResponse(w, "Erro ao criar livro", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) GetAllBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Support both header and query parameters for pagination
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = r.Header.Get("X-Page")
	}
	
	perPageStr := r.URL.Query().Get("per_page")
	if perPageStr == "" {
		perPageStr = r.Header.Get("X-Per-Page")
	}
	
	page := 1
	perPage := 10
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 {
			perPage = pp
		}
	}
	var total int
	err := h.db.QueryRow("SELECT COUNT(*) FROM livros").Scan(&total)
	if err != nil {
		log.Printf("Erro ao contar livros: %v", err)
		sendErrorResponse(w, "Erro ao buscar livros", http.StatusInternalServerError)
		return
	}
	offset := (page - 1) * perPage
	query := `
        SELECT id, name, quantity, genre_id 
        FROM livros 
        ORDER BY id 
        LIMIT $1 OFFSET $2`
	rows, err := h.db.Query(query, perPage, offset)
	if err != nil {
		log.Printf("Erro ao buscar livros: %v", err)
		sendErrorResponse(w, "Erro ao buscar livros", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var books []models.Book
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.ID, &book.Name, &book.Quantity, &book.GenreID); err != nil {
			log.Printf("Erro ao ler dados do livro: %v", err)
			sendErrorResponse(w, "Erro ao ler dados dos livros", http.StatusInternalServerError)
			return
		}
		// Adiciona o campo title para compatibilidade com frontend
		book.Title = book.Name
		books = append(books, book)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Erro após leitura dos dados: %v", err)
		sendErrorResponse(w, "Erro ao processar dados dos livros", http.StatusInternalServerError)
		return
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
	
	id := chi.URLParam(r, "id")
	
	// Se não tiver na URL, tenta extrair do corpo
	if id == "" {
		defer r.Body.Close()
		var req IDRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Erro ao decodificar JSON: %v", err)
			sendErrorResponse(w, "Erro ao ler dados", http.StatusBadRequest)
			return
		}
		id = req.ID
	}
	
	if id == "" {
		sendErrorResponse(w, "ID não fornecido", http.StatusBadRequest)
		return
	}
	
	var book models.Book
	err := h.db.QueryRow("SELECT id, name, quantity, genre_id FROM livros WHERE id = $1", id).
		Scan(&book.ID, &book.Name, &book.Quantity, &book.GenreID)
	if err == sql.ErrNoRows {
		sendErrorResponse(w, "Livro não encontrado", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Erro ao buscar livro: %v", err)
		sendErrorResponse(w, "Erro ao buscar livro", http.StatusInternalServerError)
		return
	}
	
	// Adiciona o campo title para compatibilidade com frontend
	book.Title = book.Name
	
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	id := chi.URLParam(r, "id")
	
	// Verifica parâmetros de consulta se não encontrado nos parâmetros da URL
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	
	// Como último recurso, tenta ler do corpo da requisição
	if id == "" {
		defer r.Body.Close()
		var req IDRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Erro ao decodificar JSON: %v", err)
			sendErrorResponse(w, "Erro ao ler dados", http.StatusBadRequest)
			return
		}
		id = req.ID
	}
	
	if id == "" {
		sendErrorResponse(w, "ID não fornecido", http.StatusBadRequest)
		return
	}
	
	result, err := h.db.Exec("DELETE FROM livros WHERE id = $1", id)
	if err != nil {
		log.Printf("Erro ao deletar livro: %v", err)
		sendErrorResponse(w, "Erro ao deletar livro", http.StatusInternalServerError)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		sendErrorResponse(w, "Livro não encontrado", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		log.Printf("Erro ao decodificar JSON: %v", err)
		sendErrorResponse(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}
	
	// Check if title is provided but name is not
	if book.Name == "" && book.Title != "" {
		book.Name = book.Title
	}
	
	id := chi.URLParam(r, "id")
	if id != "" {
		book.ID = id
	}
	
	if book.ID == "" {
		sendErrorResponse(w, "ID não fornecido", http.StatusBadRequest)
		return
	}
	if book.Name == "" || book.Quantity < 0 {
		sendErrorResponse(w, "Nome vazio ou quantidade negativa não permitidos", http.StatusBadRequest)
		return
	}
	
	// Verifica se o gênero existe, se fornecido
	if book.GenreID != nil {
		var count int
		err := h.db.QueryRow("SELECT COUNT(*) FROM genres WHERE id = $1", book.GenreID).Scan(&count)
		if err != nil || count == 0 {
			sendErrorResponse(w, "Gênero não encontrado", http.StatusBadRequest)
			return
		}
	}
	
	query := `UPDATE livros SET name = $1, quantity = $2, genre_id = $3 WHERE id = $4`
	result, err := h.db.Exec(query, book.Name, book.Quantity, book.GenreID, book.ID)
	if err != nil {
		log.Printf("Erro ao atualizar livro: %v", err)
		sendErrorResponse(w, "Erro ao atualizar livro", http.StatusInternalServerError)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		sendErrorResponse(w, "Livro não encontrado", http.StatusNotFound)
		return
	}
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
		sendErrorResponse(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}
	var createdBooks []models.Book
	for _, book := range books {
		// Check if title is provided but name is not
		if book.Name == "" && book.Title != "" {
			book.Name = book.Title
		}
		
		if book.Name == "" || book.Quantity < 0 {
			sendErrorResponse(w, "Nome vazio ou quantidade negativa não permitidos", http.StatusBadRequest)
			return
		}
		
		// Verifica se o gênero existe, se fornecido
		if book.GenreID != nil {
			var count int
			err := h.db.QueryRow("SELECT COUNT(*) FROM genres WHERE id = $1", book.GenreID).Scan(&count)
			if err != nil || count == 0 {
				sendErrorResponse(w, "Gênero não encontrado", http.StatusBadRequest)
				return
			}
		}
		
		newID := ksuid.New().String()
		query := `INSERT INTO livros (id, name, quantity, genre_id) VALUES ($1, $2, $3, $4) RETURNING id`
		var returnedID string
		if err := h.db.QueryRow(query, newID, book.Name, book.Quantity, book.GenreID).Scan(&returnedID); err != nil {
			log.Printf("Erro ao inserir livro: %v", err)
			sendErrorResponse(w, "Erro ao criar livro", http.StatusInternalServerError)
			return
		}
		book.ID = returnedID
		// Adiciona o campo title para compatibilidade com frontend
		book.Title = book.Name
		createdBooks = append(createdBooks, book)
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdBooks)
}
