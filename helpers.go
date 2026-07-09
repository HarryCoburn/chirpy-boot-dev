package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/HarryCoburn/chirpy-boot-dev/internal/database"
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

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithError(w http.ResponseWriter, code int, logMsg, msg string) {
	log.Println(logMsg)
	respondWithJSON(w, code, errReturn{Error: msg})
}

func toChirpResponse(chirp database.Chirp) chirpResponse {
	return chirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
}
