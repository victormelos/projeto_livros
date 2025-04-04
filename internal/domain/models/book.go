package models

// Book representa um livro no sistema
type Book struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Title    string  `json:"title,omitempty"` // Mantido para compatibilidade com o frontend
	Author   string  `json:"author"`
	Quantity int     `json:"quantity"` // Garantir que é tratado como um único valor
	GenreID  *string `json:"genre_id,omitempty"`
}
