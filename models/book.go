package models
type Book struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	GenreID  *string `json:"genre_id,omitempty"`
	Title    string  `json:"title"`
	Author   string  `json:"author"`
}
