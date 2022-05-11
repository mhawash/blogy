package main

import (
	"log"
	"net/http"
)

func main() {
	// URL Patterns
	http.HandleFunc("/api/posts", postsHandler)
	http.HandleFunc("/api/ping", pingHandler)

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
