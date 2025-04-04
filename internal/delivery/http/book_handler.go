package http

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"projeto_livros/internal/domain/models"
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
		sendErrorResponse(w, "Erro ao ler dados do livro", http.StatusBadRequest)
		return
	}

	// Usar Title como fallback para Name se necessário
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

	log.Printf("Tentando criar livro: %s, Autor: %s", book.Name, book.Author)

	// Logging detalhado para debug
	log.Printf("DEBUG - Criando livro - Título: %s", book.Name)
	// Adicionar no handler CreateBook
	log.Printf("DEBUG - Quantidade recebida: %v (tipo: %T)", book.Quantity, book.Quantity)
	log.Printf("DEBUG - Quantidade recebida para criação: %d (tipo: %T)", book.Quantity, book.Quantity)

	// Incluir a coluna author na consulta SQL
	query := `INSERT INTO livros (id, name, quantity, genre_id, author) VALUES ($1, $2, $3, $4, $5) RETURNING id`

	log.Printf("Executando consulta SQL: %s", query)

	var returnedID string
	if err := h.db.QueryRow(query, book.ID, book.Name, book.Quantity, book.GenreID, book.Author).Scan(&returnedID); err != nil {
		log.Printf("Erro ao inserir livro: %v", err)
		sendErrorResponse(w, "Erro ao criar livro", http.StatusInternalServerError)
		return
	}

	// Não precisamos mais definir book.Title = book.Name aqui
	// pois o frontend espera o campo Title

	log.Printf("Livro criado com sucesso: %s, ID: %s, Autor: %s", book.Name, book.ID, book.Author)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) GetAllBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Obter parâmetros de paginação
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("per_page")

	// Parâmetros de ordenação
	sortField := r.URL.Query().Get("sort_field")
	sortDirection := r.URL.Query().Get("sort_direction")

	// Valores padrão
	page := 1
	perPage := 20

	// Converter para inteiros se fornecidos
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

	// Calcular offset para paginação
	offset := (page - 1) * perPage

	// Construir cláusula de ordenação
	orderClause := "name ASC"
	if sortField != "" {
		// Validar campo de ordenação para evitar injeção SQL
		validFields := map[string]bool{"name": true, "quantity": true}
		if validFields[sortField] {
			orderClause = sortField
			if sortDirection == "desc" {
				orderClause += " DESC"
			} else {
				orderClause += " ASC"
			}
		}
	}

	log.Printf("Buscando livros com ordenação: %s", orderClause)

	// Incluir a coluna author na consulta SQL
	query := fmt.Sprintf(`
		SELECT id, name, quantity, genre_id, author 
		FROM livros 
		ORDER BY %s 
		LIMIT $1 OFFSET $2`, orderClause)

	log.Printf("Executando consulta SQL: %s", query)

	// Executar a consulta
	rows, err := h.db.Query(query, perPage, offset)
	if err != nil {
		log.Printf("Erro ao buscar livros: %v", err)
		sendErrorResponse(w, "Erro ao buscar livros", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Processar os resultados
	var books []models.Book
	for rows.Next() {
		var book models.Book
		var authorNull sql.NullString

		if err := rows.Scan(&book.ID, &book.Name, &book.Quantity, &book.GenreID, &authorNull); err != nil {
			log.Printf("Erro ao ler dados do livro: %v", err)
			sendErrorResponse(w, "Erro ao ler dados dos livros", http.StatusInternalServerError)
			return
		}

		// Converter author de NullString para string
		if authorNull.Valid {
			book.Author = authorNull.String
		} else {
			book.Author = ""
		}

		// Definir Title igual a Name para compatibilidade com o frontend
		book.Title = book.Name

		books = append(books, book)
	}

	// Verificar erros após a iteração
	if err := rows.Err(); err != nil {
		log.Printf("Erro ao iterar sobre os resultados: %v", err)
		sendErrorResponse(w, "Erro ao processar dados dos livros", http.StatusInternalServerError)
		return
	}

	// Contar o total de livros para paginação
	var totalBooks int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM livros").Scan(&totalBooks); err != nil {
		log.Printf("Erro ao contar livros: %v", err)
		sendErrorResponse(w, "Erro ao contar livros", http.StatusInternalServerError)
		return
	}

	// Calcular o total de páginas
	totalPages := (totalBooks + perPage - 1) / perPage

	// Construir a resposta
	response := map[string]interface{}{
		"data":        books,
		"page":        page,
		"per_page":    perPage,
		"total_books": totalBooks,
		"total_pages": totalPages,
	}

	// Log detalhado para depuração
	booksLog, _ := json.Marshal(books)
	log.Printf("Dados dos livros (formato JSON): %s", string(booksLog))
	log.Printf("Total de livros encontrados: %d", len(books))

	for i, book := range books {
		log.Printf("Livro #%d: ID=%s, Nome=%s, Autor=%s, Quantidade=%d",
			i+1, book.ID, book.Name, book.Author, book.Quantity)
	}

	log.Printf("Retornando %d livros com sucesso", len(books))

	// Enviar a resposta
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Obter ID do livro da URL ou parâmetros de consulta
	var id string

	// Verificar se estamos usando o novo padrão RESTful
	vars := chi.URLParam(r, "id")
	if vars != "" {
		id = vars
	} else {
		// Fallback para o antigo método usando parâmetros de consulta
		id = r.URL.Query().Get("id")
	}

	if id == "" {
		sendErrorResponse(w, "ID do livro não fornecido", http.StatusBadRequest)
		return
	}

	log.Printf("Buscando livro com ID: %s", id)

	// Incluir a coluna author na consulta SQL
	query := "SELECT id, name, quantity, genre_id, author FROM livros WHERE id = $1"

	log.Printf("Executando consulta SQL: %s", query)

	var book models.Book
	var authorNull sql.NullString
	err := h.db.QueryRow(query, id).Scan(&book.ID, &book.Name, &book.Quantity, &book.GenreID, &authorNull)

	if err == sql.ErrNoRows {
		sendErrorResponse(w, "Livro não encontrado", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Erro ao buscar livro: %v", err)
		sendErrorResponse(w, "Erro ao buscar livro", http.StatusInternalServerError)
		return
	}

	// Converter author de NullString para string
	if authorNull.Valid {
		book.Author = authorNull.String
	} else {
		book.Author = ""
	}

	// Garantir que title seja igual a name para compatibilidade
	book.Title = book.Name

	log.Printf("Livro encontrado com sucesso: %s", book.Name)

	w.WriteHeader(http.StatusOK)
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

// No método UpdateBook
func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		log.Printf("Erro ao decodificar JSON: %v", err)
		sendErrorResponse(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}

	// Log detalhado para debug
	log.Printf("DEBUG - Livro recebido para atualização - ID: %s, Nome: %s", book.ID, book.Name)
	log.Printf("DEBUG - Quantidade recebida no payload JSON: %d (tipo: %T)", book.Quantity, book.Quantity)

	// Verificar se o ID foi fornecido
	if book.ID == "" {
		sendErrorResponse(w, "ID do livro não fornecido", http.StatusBadRequest)
		return
	}

	// Verificar se o livro existe
	var exists bool
	if err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM livros WHERE id = $1)", book.ID).Scan(&exists); err != nil {
		log.Printf("Erro ao verificar existência do livro: %v", err)
		sendErrorResponse(w, "Erro ao verificar existência do livro", http.StatusInternalServerError)
		return
	}

	if !exists {
		sendErrorResponse(w, "Livro não encontrado", http.StatusNotFound)
		return
	}

	// Verificar se o gênero existe, se fornecido
	if book.GenreID != nil {
		var genreExists bool
		if err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM genres WHERE id = $1)", *book.GenreID).Scan(&genreExists); err != nil {
			log.Printf("Erro ao verificar existência do gênero: %v", err)
			sendErrorResponse(w, "Erro ao verificar existência do gênero", http.StatusInternalServerError)
			return
		}

		if !genreExists {
			sendErrorResponse(w, "Gênero não encontrado", http.StatusBadRequest)
			return
		}
	}

	log.Printf("Atualizando livro: %s, ID: %s, Quantidade: %d", book.Name, book.ID, book.Quantity)

	// Usar apenas o campo Name para atualização
	query := `UPDATE livros SET name = $1, quantity = $2, genre_id = $3, author = $4 WHERE id = $5`

	// Adicionar log para debug
	log.Printf("DEBUG - Quantidade sendo enviada para o banco de dados: %d", book.Quantity)

	result, err := h.db.Exec(query, book.Name, book.Quantity, book.GenreID, book.Author, book.ID)
	if err != nil {
		log.Printf("Erro ao atualizar livro: %v", err)
		sendErrorResponse(w, "Erro ao atualizar livro", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Erro ao obter linhas afetadas: %v", err)
		sendErrorResponse(w, "Erro ao atualizar livro", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		sendErrorResponse(w, "Nenhum livro foi atualizado", http.StatusNotFound)
		return
	}

	// Definir Title igual a Name para compatibilidade com o frontend
	book.Title = book.Name

	log.Printf("Livro atualizado com sucesso: %s", book.Name)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) CreateAllBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var books []models.Book
	if err := json.NewDecoder(r.Body).Decode(&books); err != nil {
		log.Printf("Erro ao decodificar JSON: %v", err)
		sendErrorResponse(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}

	createdBooks := []models.Book{}
	for _, book := range books {
		// Check if title is provided but name is not
		if book.Name == "" && book.Title != "" {
			book.Name = book.Title
		}

		if book.Name == "" || book.Quantity < 0 {
			continue
		}

		if book.GenreID != nil {
			var count int
			err := h.db.QueryRow("SELECT COUNT(*) FROM genres WHERE id = $1", book.GenreID).Scan(&count)
			if err != nil || count == 0 {
				continue
			}
		}

		newID := ksuid.New().String()

		log.Printf("Tentando criar livro em lote: %s, Autor: %s", book.Name, book.Author)

		// Preparar o valor do autor (que pode ser NULL)
		var authorValue interface{}
		if book.Author == "" {
			authorValue = nil
		} else {
			authorValue = book.Author
		}

		// Incluir a coluna author na consulta SQL
		query := `INSERT INTO livros (id, name, quantity, genre_id, author) VALUES ($1, $2, $3, $4, $5) RETURNING id`

		var returnedID string
		if err := h.db.QueryRow(query, newID, book.Name, book.Quantity, book.GenreID, authorValue).Scan(&returnedID); err != nil {
			log.Printf("Erro ao inserir livro: %v", err)
			continue
		}

		book.ID = newID
		// Garantir que title seja igual a name para compatibilidade
		book.Title = book.Name
		createdBooks = append(createdBooks, book)
	}

	response := map[string]interface{}{
		"message": fmt.Sprintf("%d livros criados com sucesso", len(createdBooks)),
		"books":   createdBooks,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
