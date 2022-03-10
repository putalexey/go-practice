package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type GZipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w GZipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// GZipEncoder middleware compresses response with gzip, if client accepts gzip encoding with "Accept-Encoding" header
func GZipEncoder(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		writer := w

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer gz.Close()
			writer = GZipWriter{ResponseWriter: w, Writer: gz}
			writer.Header().Set("Content-Encoding", "gzip")
		}
		next.ServeHTTP(writer, r)
	}
	return http.HandlerFunc(fn)
}
