package urlgenerator

import (
	"fmt"
	"strconv"
)

type SequenceGenerator struct {
	Domain  string
	counter int64
}

func (g *SequenceGenerator) GenerateShort(_ string) string {
	str := strconv.FormatInt(g.counter, 36)
	g.counter += 1
	return str
}

func (g *SequenceGenerator) GetURL(short string) string {
	return fmt.Sprintf("http://%s/%s", g.Domain, short)
}
