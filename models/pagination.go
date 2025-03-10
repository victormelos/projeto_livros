package models
type PaginationRequest struct {
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Sort    string `json:"sort"`
	Order   string `json:"order"`
}
type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Pagination struct {
		CurrentPage int  `json:"current_page"`
		PerPage     int  `json:"per_page"`
		TotalItems  int  `json:"total_items"`
		TotalPages  int  `json:"total_pages"`
		HasPrevious bool `json:"has_previous"`
		HasNext     bool `json:"has_next"`
	} `json:"pagination"`
}
type PaginationMeta struct {
	Links struct {
		Self     string `json:"self"`
		First    string `json:"first"`
		Previous string `json:"previous,omitempty"`
		Next     string `json:"next,omitempty"`
		Last     string `json:"last"`
	} `json:"links"`
}
