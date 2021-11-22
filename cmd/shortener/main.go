package main

import (
	"github.com/putalexey/go-practicum/internal/app/shortener"
	"log"
	"net/http"
)

var domain = "localhost:8080"

func main() {
	handler := shortener.NewShortener(domain, nil)
	log.Fatal(http.ListenAndServe(":8080", handler))
}
