package responses

type CreateShortResponse struct {
	Result string `json:"result"`
}

type ListShortItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type ListShortsResponse []ListShortItem

type ErrorResponse struct {
	Error string `json:"error"`
}
