package responses

type CreateShortResponse struct {
	Result string `json:"result"`
}

type ListShortItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type ListShortsResponse []ListShortItem

type CreateShortBatchResponse []CreateShortBatchResponseItem

type CreateShortBatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
