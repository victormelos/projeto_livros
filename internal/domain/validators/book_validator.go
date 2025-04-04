package validators

import (
	"encoding/json"
	"net/http"
	"projeto_livros/internal/domain/errors"
	"projeto_livros/internal/domain/models"

	"strings"
)

func ValidateBookInput(r *http.Request) (*models.Book, error) {
	defer r.Body.Close()
	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		return nil, errors.NewBadRequestError("Formato de JSON inválido: " + err.Error())
	}
	trimmedTitle := strings.TrimSpace(book.Title)
	if trimmedTitle == "" {
		return nil, errors.NewBadRequestError("O campo 'title' é obrigatório")
	}
	book.Title = trimmedTitle
	trimmedAuthor := strings.TrimSpace(book.Author)
	if trimmedAuthor == "" {
		return nil, errors.NewBadRequestError("O campo 'author' é obrigatório")
	}
	book.Author = trimmedAuthor
	if book.Quantity < 0 {
		return nil, errors.NewBadRequestError("A quantidade não pode ser negativa")
	}
	return &book, nil
}
