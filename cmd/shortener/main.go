package main

import (
	"github.com/putalexey/go-practicum/internal/app/shortener"
	"log"
	"net/http"
)

var domain = "localhost:8080"

func main() {
	handler := shortener.NewShortener(domain)

	http.Handle("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
