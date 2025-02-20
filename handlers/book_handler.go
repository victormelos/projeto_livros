package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"projeto_livros/models"

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
	defer r.Body.Close()

	// Gerar ID único usando ksuid
	book.ID = ksuid.New().String()

	query := `INSERT INTO livros (id, name, quantity) VALUES ($1, $2, $3) RETURNING id`
	var returnedID string
	if err := h.db.QueryRow(query, book.ID, book.Name, book.Quantity).Scan(&returnedID); err != nil {
		log.Printf("Erro ao inserir livro: %v", err)
		http.Error(w, "Erro ao criar livro", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) GetAllBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := h.db.Ping(); err != nil {
		log.Printf("Erro de conexão com o banco: %v", err)
		http.Error(w, "Erro de conexão com o banco de dados", http.StatusInternalServerError)
		return
	}

	log.Println("Iniciando busca de livros...")
	rows, err := h.db.Query("SELECT id, name, quantity FROM livros")
	if err != nil {
		log.Printf("Erro ao executar query: %v", err)
		http.Error(w, "Erro ao buscar livros", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.ID, &book.Name, &book.Quantity); err != nil {
			log.Printf("Erro ao ler dados do livro: %v", err)
			http.Error(w, "Erro ao ler dados dos livros", http.StatusInternalServerError)
			return
		}
		books = append(books, book)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Erro após leitura dos dados: %v", err)
		http.Error(w, "Erro ao processar dados dos livros", http.StatusInternalServerError)
		return
	}

	if books == nil {
		books = []models.Book{}
	}

	log.Printf("Encontrados %d livros", len(books))
	if err := json.NewEncoder(w).Encode(books); err != nil {
		log.Printf("Erro ao codificar resposta JSON: %v", err)
		http.Error(w, "Erro ao gerar resposta", http.StatusInternalServerError)
		return
	}
}

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req IDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Erro ao decodificar JSON: %v", err)
		http.Error(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

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

}
