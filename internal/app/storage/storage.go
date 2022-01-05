package storage

type Record struct {
	Short  string `json:"short"`
	Full   string `json:"full"`
	UserID string `json:"user_id"`
}

type RecordMap map[string]Record

type Storager interface {
	Store(short, full, userID string) error
	Load(short string) (string, error)
	LoadForUser(userID string) ([]Record, error)
	Delete(short string) error
}
