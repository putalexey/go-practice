package responses

type CreateShortResponse struct {
	Result string `json:"result" example:"http://shortener.org/123"`
}

type ListShortItem struct {
	ShortURL    string `json:"short_url" example:"http://shortener.org/123"`
	OriginalURL string `json:"original_url" example:"http://example.com/"`
}

type ListShortsResponse []ListShortItem

type CreateShortBatchResponse []CreateShortBatchResponseItem

type CreateShortBatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"Not found"`
}
