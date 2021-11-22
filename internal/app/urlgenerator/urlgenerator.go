package urlgenerator

type URLGenerator interface {
	GenerateShort(fullURL string) string
	GetURL(short string) string
}
