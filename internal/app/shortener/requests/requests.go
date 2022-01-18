package requests

type CreateShortRequest struct {
	URL string `json:"url"`
}

type CreateShortBatchRequest []CreateShortBatchItem

type CreateShortBatchItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}
