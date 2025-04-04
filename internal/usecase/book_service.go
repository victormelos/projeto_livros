package services

import (
	"database/sql"
	"log"
	"projeto_livros/internal/domain/errors"
	"projeto_livros/internal/domain/models"

	"github.com/segmentio/ksuid"
)

type BookService interface {
	CreateBook(book *models.Book) error
	GetAllBooks(page, perPage int) ([]models.Book, int, error)
	GetBookByID(id string) (*models.Book, error)
	UpdateBook(book *models.Book) error
	DeleteBook(id string) error
}
type BookServiceImpl struct {
	db *sql.DB
}

func NewBookService(db *sql.DB) BookService {
	return &BookServiceImpl{db: db}
}

// No método CreateBook
func (s *BookServiceImpl) CreateBook(book *models.Book) error {
	if book.Title == "" {
		return errors.NewBadRequestError("O título do livro é obrigatório")
	}
	if book.Author == "" {
		return errors.NewBadRequestError("O autor do livro é obrigatório")
	}
	// Garantir que a quantidade é um número válido
	if book.Quantity < 0 {
		return errors.NewBadRequestError("A quantidade não pode ser negativa")
	}

	// Adicionar log para debug
	log.Printf("DEBUG - Quantidade recebida para criação no serviço: %d", book.Quantity)

	if book.GenreID != nil {
		var count int
		err := s.db.QueryRow("SELECT COUNT(*) FROM genres WHERE id = $1", book.GenreID).Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			return errors.NewBadRequestError("Gênero não encontrado")
		}
	}
	book.ID = ksuid.New().String()

	// Modificar a query para usar apenas os campos necessários
	query := `INSERT INTO livros (id, name, author, quantity, genre_id) 
              VALUES ($1, $2, $3, $4, $5)`

	// Adicionar log para debug
	log.Printf("DEBUG - Quantidade sendo enviada para o banco de dados no serviço: %d", book.Quantity)

	_, err := s.db.Exec(query, book.ID, book.Name, book.Author, book.Quantity, book.GenreID)
	return err
}
func (s *BookServiceImpl) GetAllBooks(page, perPage int) ([]models.Book, int, error) {
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM livros").Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * perPage
	query := `
        SELECT id, name, author, quantity, genre_id 
        FROM livros 
        ORDER BY id 
        LIMIT $1 OFFSET $2`
	rows, err := s.db.Query(query, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var books []models.Book
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.ID, &book.Name, &book.Author, &book.Quantity, &book.GenreID); err != nil {
			return nil, 0, err
		}
		// Como title não existe na tabela, vamos usar o valor de name
		book.Title = book.Name
		books = append(books, book)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	return books, total, nil
}
func (s *BookServiceImpl) GetBookByID(id string) (*models.Book, error) {
	if id == "" {
		return nil, errors.NewBadRequestError("ID não fornecido")
	}
	var book models.Book
	err := s.db.QueryRow("SELECT id, name, author, quantity, genre_id FROM livros WHERE id = $1", id).
		Scan(&book.ID, &book.Name, &book.Author, &book.Quantity, &book.GenreID)

	// Como title não existe na tabela, vamos usar o valor de name
	book.Title = book.Name
	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("Livro não encontrado")
	} else if err != nil {
		return nil, err
	}
	return &book, nil
}

// No método UpdateBook
func (s *BookServiceImpl) UpdateBook(book *models.Book) error {
	if book.ID == "" {
		return errors.NewBadRequestError("ID não fornecido")
	}
	if book.Title == "" {
		return errors.NewBadRequestError("O título do livro é obrigatório")
	}
	if book.Author == "" {
		return errors.NewBadRequestError("O autor do livro é obrigatório")
	}
	if book.Quantity < 0 {
		return errors.NewBadRequestError("A quantidade não pode ser negativa")
	}
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM livros WHERE id = $1)", book.ID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return errors.NewNotFoundError("Livro não encontrado")
	}

	// Modificar a query para usar apenas os campos necessários
	query := `UPDATE livros 
              SET name = $1, author = $2, quantity = $3, genre_id = $4 
              WHERE id = $5`

	// Adicionar log para debug
	log.Printf("DEBUG - Quantidade recebida para atualização no serviço: %d", book.Quantity)

	// Usar os valores corretos para cada campo
	_, err = s.db.Exec(query, book.Name, book.Author, book.Quantity, book.GenreID, book.ID)
	return err
}
func (s *BookServiceImpl) DeleteBook(id string) error {
	if id == "" {
		return errors.NewBadRequestError("ID não fornecido")
	}
	result, err := s.db.Exec("DELETE FROM livros WHERE id = $1", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.NewNotFoundError("Livro não encontrado")
	}
	return nil
}
