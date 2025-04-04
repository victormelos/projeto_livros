package models
type Genre struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
type GenreWithBooks struct {
	Name       string `json:"name"`
	TotalBooks int    `json:"total_books"`
	Books      []Book `json:"books"`
}
