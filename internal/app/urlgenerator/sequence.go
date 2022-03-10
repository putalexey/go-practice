package urlgenerator

import (
	"fmt"
	"strconv"
)

// SequenceGenerator generates ids for the short urls. Generates sequence like: 0, 1, 2, 3...
type SequenceGenerator struct {
	BaseURL string
	counter int64
}

// GenerateShort get next id for the short url
func (g *SequenceGenerator) GenerateShort(_ string) string {
	str := strconv.FormatInt(g.counter, 36)
	g.counter += 1
	return str
}

// GetURL returns absolute url, combining SequenceGenerator.BaseURL with short
func (g *SequenceGenerator) GetURL(short string) string {
	return fmt.Sprintf("%s/%s", g.BaseURL, short)
}
