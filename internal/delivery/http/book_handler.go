package http

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"projeto_livros/internal/domain/models"
	"strconv"
	"strings"

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

	if book.Name == "" || book.Quantity <= 0 {
		sendErrorResponse(w, "Nome vazio ou quantidade inválida. A quantidade deve ser maior que zero.", http.StatusBadRequest)
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

	log.Printf("DEBUG - LISTAGEM: Buscando livros com ordenação: %s", orderClause)

	// IMPORTANTE: Buscar livros SEM MODIFICAR os valores originais
	query := fmt.Sprintf(`
		SELECT id, name, quantity, genre_id, author 
		FROM livros 
		ORDER BY %s 
		LIMIT $1 OFFSET $2`, orderClause)

	log.Printf("DEBUG - LISTAGEM: Executando query SQL: %s", query)

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

		// Log para cada linha lida do banco
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

		// Log detalhado para cada livro encontrado na busca
		log.Printf("DEBUG - LISTAGEM: Livro encontrado - ID: %s, Nome: %s, Quantidade (direto do banco): %d",
			book.ID, book.Name, book.Quantity)

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
	for i, book := range books {
		log.Printf("DEBUG - LISTAGEM: Livro #%d (antes de enviar resposta): ID=%s, Nome=%s, Autor=%s, Quantidade=%d",
			i+1, book.ID, book.Name, book.Author, book.Quantity)
	}

	log.Printf("DEBUG - LISTAGEM: Retornando %d livros", len(books))

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

	// Limpar o ID para garantir que não tenha caracteres extras
	parts := strings.Split(id, ":")
	id = parts[0] // Pegar apenas a parte antes de : (se houver)

	log.Printf("Buscando livro com ID (limpo): %s", id)

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

// UpdateBook atualiza um livro existente
func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Ler o corpo da requisição
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Erro ao ler corpo da requisição: %v", err)
		sendErrorResponse(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Debugging - imprimir o corpo completo da requisição
	log.Printf("DEBUG - CORPO COMPLETO da requisição de atualização: %s", string(bodyBytes))

	// Primeiro extrair ID do livro da URL
	var bookID string
	// Tentar obter da URL RESTful primeiro
	if chi.URLParam(r, "id") != "" {
		bookID = chi.URLParam(r, "id")
	} else {
		// Se não encontrar na URL, verificar no corpo da requisição
		var requestData map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
			log.Printf("Erro ao decodificar JSON: %v", err)
			sendErrorResponse(w, "Erro ao ler dados", http.StatusBadRequest)
			return
		}

		// Verificar se o ID está no corpo
		if id, ok := requestData["id"].(string); ok {
			bookID = id
		} else {
			sendErrorResponse(w, "ID do livro não fornecido", http.StatusBadRequest)
			return
		}
	}

	// Verificar se o payload é apenas para atualização de quantidade
	var requestData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &requestData); err != nil {
		log.Printf("Erro ao decodificar JSON para map: %v", err)
		sendErrorResponse(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}

	quantityOnly := false

	// Verificar se o payload contém apenas id e quantity
	if len(requestData) == 2 && requestData["id"] != nil && requestData["quantity"] != nil {
		quantityOnly = true
	}

	// Caso especial para atualizações apenas de quantidade
	if quantityOnly {
		log.Printf("DEBUG - Detectada atualização apenas de quantidade para o livro: %s", bookID)

		quantityValue, ok := requestData["quantity"].(float64)
		if !ok {
			log.Printf("DEBUG - Quantidade inválida no payload: %v (tipo: %T)", requestData["quantity"], requestData["quantity"])
			sendErrorResponse(w, "Formato de quantidade inválido", http.StatusBadRequest)
			return
		}

		log.Printf("DEBUG - QUANTIDADE: Valor float64 da requisição: %v", quantityValue)

		// Conversão para int sem manipulação
		newQuantity := int(quantityValue)
		log.Printf("DEBUG - QUANTIDADE: Após conversão para int: %d", newQuantity)

		log.Printf("DEBUG - Atualizando apenas quantidade - ID: %s, Nova quantidade: %d", bookID, newQuantity)

		// Verificar se o livro existe
		var exists bool
		if err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM livros WHERE id = $1)", bookID).Scan(&exists); err != nil {
			log.Printf("Erro ao verificar existência do livro: %v", err)
			sendErrorResponse(w, "Erro ao verificar existência do livro", http.StatusInternalServerError)
			return
		}

		if !exists {
			sendErrorResponse(w, "Livro não encontrado", http.StatusNotFound)
			return
		}

		// Obter quantidade atual para debug
		var currentQuantity int
		if err := h.db.QueryRow("SELECT quantity FROM livros WHERE id = $1", bookID).Scan(&currentQuantity); err == nil {
			log.Printf("DEBUG - Quantidade atual no banco antes de atualizar: %d, Nova quantidade: %d", currentQuantity, newQuantity)
		}

		// Validar nova quantidade
		if newQuantity <= 0 {
			log.Printf("DEBUG - QUANTIDADE: Rejeitada por ser <= 0: %d", newQuantity)
			sendErrorResponse(w, "A quantidade deve ser maior que zero", http.StatusBadRequest)
			return
		}

		// Executar atualização direta da quantidade - IMPORTANTE: sem nenhuma manipulação
		query := "UPDATE livros SET quantity = $1 WHERE id = $2"

		log.Printf("DEBUG - Executando query direta para quantidade: %s - Params: [%d, %s]", query, newQuantity, bookID)

		// IMPORTANTE: Verificar tipo da coluna quantity no banco
		var columnType string
		err = h.db.QueryRow("SELECT data_type FROM information_schema.columns WHERE table_name = 'livros' AND column_name = 'quantity'").Scan(&columnType)
		if err == nil {
			log.Printf("DEBUG - QUANTIDADE: Tipo da coluna no banco: %s", columnType)
		} else {
			log.Printf("DEBUG - QUANTIDADE: Erro ao verificar tipo da coluna: %v", err)
		}

		result, err := h.db.Exec(query, newQuantity, bookID)
		if err != nil {
			log.Printf("Erro ao atualizar quantidade: %v", err)
			sendErrorResponse(w, "Erro ao atualizar quantidade", http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			sendErrorResponse(w, "Nenhum livro foi atualizado", http.StatusNotFound)
			return
		}

		// Verificar quantidade após atualização
		var finalQuantity int
		if err := h.db.QueryRow("SELECT quantity FROM livros WHERE id = $1", bookID).Scan(&finalQuantity); err != nil {
			log.Printf("Erro ao verificar quantidade final: %v", err)
		} else {
			log.Printf("DEBUG - QUANTIDADE: Quantidade após atualização diretamente do banco: %d", finalQuantity)
			log.Printf("DEBUG - ANÁLISE DETALHADA: Quantidade solicitada: %d, Quantidade final no banco: %d, Razão: %.4f",
				newQuantity, finalQuantity, float64(finalQuantity)/float64(newQuantity))

			// Verificar se a quantidade no banco é diferente da solicitada
			if finalQuantity != newQuantity {
				log.Printf("DEBUG - ALERTA: Quantidade no banco (%d) diferente da solicitada (%d)!", finalQuantity, newQuantity)

				// Tentar novamente com uma abordagem diferente, usando prepared statement
				stmt, err := h.db.Prepare("UPDATE livros SET quantity = $1 WHERE id = $2")
				if err != nil {
					log.Printf("DEBUG - Erro ao preparar statement: %v", err)
				} else {
					defer stmt.Close()
					_, err = stmt.Exec(newQuantity, bookID)
					if err != nil {
						log.Printf("DEBUG - Erro na segunda tentativa com prepared statement: %v", err)
					} else {
						// Verificar se a segunda tentativa funcionou
						h.db.QueryRow("SELECT quantity FROM livros WHERE id = $1", bookID).Scan(&finalQuantity)
						log.Printf("DEBUG - QUANTIDADE: Após segunda tentativa: %d", finalQuantity)
					}
				}
			}
		}

		// Retornar resposta com informações da atualização
		response := struct {
			ID           string `json:"id"`
			RequestedQty int    `json:"requested_quantity"`
			FinalQty     int    `json:"final_quantity"`
			Success      bool   `json:"success"`
			Message      string `json:"message,omitempty"`
		}{
			ID:           bookID,
			RequestedQty: newQuantity,
			FinalQty:     finalQuantity,
			Success:      finalQuantity == newQuantity,
		}

		if finalQuantity != newQuantity {
			response.Message = fmt.Sprintf("A quantidade foi atualizada, mas com um valor diferente do solicitado. Solicitado: %d, Final: %d",
				newQuantity, finalQuantity)
		} else {
			response.Message = "Quantidade atualizada com sucesso"
		}

		// Verificar novamente o valor no banco antes de enviar a resposta
		var checkQuantity int
		if err := h.db.QueryRow("SELECT quantity FROM livros WHERE id = $1", bookID).Scan(&checkQuantity); err == nil {
			log.Printf("DEBUG - QUANTIDADE: Verificação final antes da resposta: %d", checkQuantity)
			if checkQuantity != finalQuantity {
				log.Printf("DEBUG - ALERTA: Valor mudou entre a atualização e a resposta! Anterior: %d, Atual: %d",
					finalQuantity, checkQuantity)
				response.FinalQty = checkQuantity
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Se chegou aqui, é uma atualização normal de livro (não apenas da quantidade)
	var book models.Book
	if err := json.NewDecoder(bytes.NewBuffer(bodyBytes)).Decode(&book); err != nil {
		log.Printf("Erro ao decodificar JSON para struct Book: %v", err)
		sendErrorResponse(w, "Erro ao ler dados", http.StatusBadRequest)
		return
	}

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
	log.Printf("DEBUG - Query de atualização completa: %s - Params: [%s, %d, %v, %s, %s]",
		query, book.Name, book.Quantity, book.GenreID, book.Author, book.ID)

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
		if book.Name == "" && book.Title != "" {
			book.Name = book.Title
		}

		if book.Name == "" || book.Quantity <= 0 {
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

// UpdateBookQuantity - Endpoint específico apenas para atualizar a quantidade
func (h *BookHandler) UpdateBookQuantity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Log detalhado da requisição
	log.Printf("ROTA ESPECIAL - UpdateBookQuantity: Método=%s, ContentType=%s",
		r.Method, r.Header.Get("Content-Type"))

	// Verificar método HTTP
	if r.Method != "POST" {
		log.Printf("ROTA ESPECIAL - Método incorreto: %s (esperado POST)", r.Method)
		sendErrorResponse(w, fmt.Sprintf("Método %s não permitido para esta rota", r.Method), http.StatusMethodNotAllowed)
		return
	}

	// Capturar o payload bruto
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ROTA ESPECIAL - Erro ao ler corpo: %v", err)
		sendErrorResponse(w, "Erro ao ler dados da requisição", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Log detalhado do payload recebido
	log.Printf("ROTA ESPECIAL - Payload bruto: %s", string(bodyBytes))

	// Extrair apenas o ID e a quantidade
	type quantityUpdate struct {
		ID       string `json:"id"`
		Quantity int    `json:"quantity"`
	}

	var update quantityUpdate
	if err := json.Unmarshal(bodyBytes, &update); err != nil {
		log.Printf("ROTA ESPECIAL - Erro ao decodificar: %v - Payload: %s", err, string(bodyBytes))
		sendErrorResponse(w, "Formato inválido. Esperado: {\"id\":\"...\", \"quantity\":N}", http.StatusBadRequest)
		return
	}

	// Validar dados recebidos
	if update.ID == "" {
		log.Printf("ROTA ESPECIAL - ID não fornecido")
		sendErrorResponse(w, "ID do livro não fornecido", http.StatusBadRequest)
		return
	}

	if update.Quantity <= 0 {
		log.Printf("ROTA ESPECIAL - Quantidade inválida: %d", update.Quantity)
		sendErrorResponse(w, "A quantidade deve ser maior que zero", http.StatusBadRequest)
		return
	}

	log.Printf("ROTA ESPECIAL - Processando: ID=%s, Quantidade=%d", update.ID, update.Quantity)

	// Verificar se o livro existe
	var exists bool
	if err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM livros WHERE id = $1)", update.ID).Scan(&exists); err != nil {
		log.Printf("ROTA ESPECIAL - Erro ao verificar existência: %v", err)
		sendErrorResponse(w, "Erro ao verificar existência do livro", http.StatusInternalServerError)
		return
	}

	if !exists {
		log.Printf("ROTA ESPECIAL - Livro não encontrado: %s", update.ID)
		sendErrorResponse(w, "Livro não encontrado", http.StatusNotFound)
		return
	}

	// Verificar quantidade atual no banco
	var currentQuantity int
	if err := h.db.QueryRow("SELECT quantity FROM livros WHERE id = $1", update.ID).Scan(&currentQuantity); err != nil {
		log.Printf("ROTA ESPECIAL - Erro ao obter quantidade atual: %v", err)
	} else {
		log.Printf("ROTA ESPECIAL - Quantidade atual no banco: %d, Nova quantidade solicitada: %d",
			currentQuantity, update.Quantity)
	}

	// Usar uma query SQL direta que APENAS atualiza a quantidade
	query := "UPDATE livros SET quantity = $1 WHERE id = $2"

	log.Printf("ROTA ESPECIAL - Executando query: %s com [%d, %s]", query, update.Quantity, update.ID)

	// Executar a query
	result, err := h.db.Exec(query, update.Quantity, update.ID)
	if err != nil {
		log.Printf("ROTA ESPECIAL - Erro na atualização: %v", err)
		sendErrorResponse(w, "Erro ao atualizar quantidade", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("ROTA ESPECIAL - Nenhuma linha afetada")
		sendErrorResponse(w, "Nenhum livro foi atualizado", http.StatusNotFound)
		return
	}

	log.Printf("ROTA ESPECIAL - Atualização concluída, %d linhas afetadas", rowsAffected)

	// Verificar a quantidade após a atualização
	var finalQuantity int
	if err := h.db.QueryRow("SELECT quantity FROM livros WHERE id = $1", update.ID).Scan(&finalQuantity); err != nil {
		log.Printf("ROTA ESPECIAL - Erro ao verificar quantidade final: %v", err)
	} else {
		log.Printf("ROTA ESPECIAL - Quantidade APÓS atualização: %d", finalQuantity)

		// Verificar se a quantidade foi atualizada corretamente
		if finalQuantity != update.Quantity {
			log.Printf("ROTA ESPECIAL - ATENÇÃO: Valor no banco (%d) diferente do solicitado (%d)",
				finalQuantity, update.Quantity)
		}
	}

	// Retornar resposta
	response := struct {
		ID           string `json:"id"`
		RequestedQty int    `json:"requested_quantity"`
		FinalQty     int    `json:"final_quantity"`
		Success      bool   `json:"success"`
		Message      string `json:"message,omitempty"`
		OriginalQty  int    `json:"original_quantity"`
	}{
		ID:           update.ID,
		RequestedQty: update.Quantity,
		FinalQty:     finalQuantity,
		Success:      finalQuantity == update.Quantity,
		OriginalQty:  currentQuantity,
	}

	if finalQuantity != update.Quantity {
		response.Message = fmt.Sprintf("Atenção: Valor no banco (%d) diferente do solicitado (%d)",
			finalQuantity, update.Quantity)
	} else {
		response.Message = "Quantidade atualizada com sucesso"
	}

	log.Printf("ROTA ESPECIAL - Enviando resposta: %+v", response)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateQuantityDirect - Método simples para atualizar a quantidade via query params
func (h *BookHandler) UpdateQuantityDirect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Obter parâmetros da consulta
	bookID := r.URL.Query().Get("id")
	quantityStr := r.URL.Query().Get("quantity")

	log.Printf("MÉTODO DIRETO - Recebido: ID=%s, Quantidade=%s", bookID, quantityStr)

	// Validar ID
	if bookID == "" {
		log.Printf("MÉTODO DIRETO - ID não fornecido")
		sendErrorResponse(w, "ID do livro não fornecido", http.StatusBadRequest)
		return
	}

	// Validar e converter quantidade
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil || quantity <= 0 {
		log.Printf("MÉTODO DIRETO - Quantidade inválida: %s", quantityStr)
		sendErrorResponse(w, "Quantidade inválida. Deve ser um número maior que zero.", http.StatusBadRequest)
		return
	}

	// Verificar se o livro existe
	var exists bool
	if err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM livros WHERE id = $1)", bookID).Scan(&exists); err != nil {
		log.Printf("MÉTODO DIRETO - Erro ao verificar livro: %v", err)
		sendErrorResponse(w, "Erro ao verificar existência do livro", http.StatusInternalServerError)
		return
	}

	if !exists {
		log.Printf("MÉTODO DIRETO - Livro não encontrado: %s", bookID)
		sendErrorResponse(w, "Livro não encontrado", http.StatusNotFound)
		return
	}

	// Obter quantidade atual para comparação
	var currentQuantity int
	if err := h.db.QueryRow("SELECT quantity FROM livros WHERE id = $1", bookID).Scan(&currentQuantity); err == nil {
		log.Printf("MÉTODO DIRETO - Quantidade atual: %d, Nova quantidade: %d", currentQuantity, quantity)
	}

	// MODIFICAÇÃO IMPORTANTE: Usar um comando SQL preparado para garantir que o valor seja exatamente como fornecido
	stmt, err := h.db.Prepare("UPDATE livros SET quantity = $1 WHERE id = $2")
	if err != nil {
		log.Printf("MÉTODO DIRETO - Erro ao preparar statement: %v", err)
		sendErrorResponse(w, "Erro interno ao preparar atualização", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// Executar a atualização com o valor exato recebido
	log.Printf("MÉTODO DIRETO - Executando statement com exatamente: [%d, %s]", quantity, bookID)
	result, err := stmt.Exec(quantity, bookID)
	if err != nil {
		log.Printf("MÉTODO DIRETO - Erro na atualização: %v", err)
		sendErrorResponse(w, "Erro ao atualizar quantidade", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("MÉTODO DIRETO - Nenhuma linha afetada")
		sendErrorResponse(w, "Nenhum livro foi atualizado", http.StatusNotFound)
		return
	}

	// Verificar quantidade final após atualização
	var finalQuantity int
	if err := h.db.QueryRow("SELECT quantity FROM livros WHERE id = $1", bookID).Scan(&finalQuantity); err != nil {
		log.Printf("MÉTODO DIRETO - Erro ao verificar quantidade final: %v", err)
	} else {
		log.Printf("MÉTODO DIRETO - Quantidade final: %d", finalQuantity)

		// Verificar se houve discrepância
		if finalQuantity != quantity {
			log.Printf("MÉTODO DIRETO - ALERTA! Valor no banco (%d) diferente do solicitado (%d). Tentando correção...",
				finalQuantity, quantity)

			// Tentar uma última vez com consulta direta
			_, directErr := h.db.Exec("UPDATE livros SET quantity = $1 WHERE id = $2", quantity, bookID)
			if directErr == nil {
				// Verificar novamente
				var checkQuantity int
				h.db.QueryRow("SELECT quantity FROM livros WHERE id = $1", bookID).Scan(&checkQuantity)
				log.Printf("MÉTODO DIRETO - Após correção final: %d", checkQuantity)
				finalQuantity = checkQuantity
			}
		}
	}

	// Retornar resposta de sucesso com dados atualizados
	response := struct {
		ID           string `json:"id"`
		Quantity     int    `json:"quantity"`
		RequestedQty int    `json:"requested_quantity"`
		OriginalQty  int    `json:"original_quantity"`
		Success      bool   `json:"success"`
		Message      string `json:"message"`
	}{
		ID:           bookID,
		Quantity:     finalQuantity,
		RequestedQty: quantity,
		OriginalQty:  currentQuantity,
		Success:      finalQuantity == quantity,
		Message:      fmt.Sprintf("Quantidade atualizada para %d", finalQuantity),
	}

	log.Printf("MÉTODO DIRETO - Resposta: %+v", response)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
