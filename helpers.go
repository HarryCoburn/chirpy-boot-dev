package main

import (
	"fmt"
	"net/http"
)

const (
	contentTypeJSON  = "application/json"
	contentTypeHTML  = "text/html; charset=utf-8"
	contentTypePlain = "text/plain; charset=utf-8"
)

func respondWith(w http.ResponseWriter, code int, contentType, body string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	fmt.Fprint(w, body)
}
