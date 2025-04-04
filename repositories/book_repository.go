package repositories

import (
	"database/sql"
	"projeto_livros/models"
)

type BookRepository interface {
	Create(book *models.Book) error
	FindAll(limit, offset int) ([]models.Book, error)
	FindByID(id string) (*models.Book, error)
	Update(book *models.Book) (int64, error)
	Delete(id string) (int64, error)
	Count() (int, error)
}
type PostgresBookRepository struct {
	db *sql.DB
}

func NewPostgresBookRepository(db *sql.DB) BookRepository {
	return &PostgresBookRepository{db: db}
}

// Remove the first implementation and keep only this one
func (r *PostgresBookRepository) Create(book *models.Book) error {
	// Remover o campo title da query, já que estamos usando apenas name
	query := `INSERT INTO livros (id, name, quantity, genre_id, author) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var returnedID string
	err := r.db.QueryRow(
		query,
		book.ID,
		book.Name,
		book.Quantity,
		book.GenreID,
		book.Author,
	).Scan(&returnedID)
	return err
}

func (r *PostgresBookRepository) FindAll(limit, offset int) ([]models.Book, error) {
	query := `
        SELECT id, name, quantity, genre_id, author
        FROM livros 
        ORDER BY id 
        LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var books []models.Book
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.ID, &book.Name, &book.Quantity, &book.GenreID, &book.Author); err != nil {
			return nil, err
		}
		// Set title equal to name for frontend compatibility
		book.Title = book.Name
		books = append(books, book)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return books, nil
}

func (r *PostgresBookRepository) FindByID(id string) (*models.Book, error) {
	query := `SELECT id, name, quantity, genre_id, author FROM livros WHERE id = $1`
	var book models.Book
	err := r.db.QueryRow(query, id).Scan(
		&book.ID,
		&book.Name,
		&book.Quantity,
		&book.GenreID,
		&book.Author,
	)
	if err != nil {
		return nil, err
	}
	// Set title equal to name for frontend compatibility
	book.Title = book.Name
	return &book, nil
}

func (r *PostgresBookRepository) Update(book *models.Book) (int64, error) {
	// Remover o campo title da query, já que estamos usando apenas name
	query := `UPDATE livros 
              SET name = $1, quantity = $2, genre_id = $3, author = $4 
              WHERE id = $5`
	result, err := r.db.Exec(
		query,
		book.Name,
		book.Quantity,
		book.GenreID,
		book.Author,
		book.ID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *PostgresBookRepository) Delete(id string) (int64, error) {
	query := `DELETE FROM livros WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *PostgresBookRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM livros").Scan(&count)
	return count, err
}
