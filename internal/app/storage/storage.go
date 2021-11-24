package storage

type Storager interface {
	Store(short, full string) error
	Load(short string) (string, error)
	Delete(short string) error
}
