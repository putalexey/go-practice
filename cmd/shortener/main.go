package main

import (
	"github.com/putalexey/go-practicum/internal/app/shortener"
	"log"
	"net/http"
)

var domain = "localhost:8080"

func main() {
	router := shortener.NewRouter(domain, nil)
	log.Fatal(http.ListenAndServe(":8080", router))
}
