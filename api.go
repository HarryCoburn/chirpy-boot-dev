package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

type parameters struct {
	Body string `json:"body"`
}

type errReturn struct {
	Error string `json:"error"`
}

type validReturn struct {
	CleanedBody string `json:"cleaned_body"`
}

type newUser struct {
	Email string `json:"email"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func servHealth(write http.ResponseWriter, request *http.Request) {
	write.Header().Set("Content-Type", "text/plain; charset=utf-8")
	write.WriteHeader(http.StatusOK)
	write.Write([]byte("OK"))
}

func chirpValidate(write http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respBody := errReturn{
			Error: "Something went wrong",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			write.WriteHeader(500)
			return
		}
		write.Header().Set("Content-Type", "application/json")
		write.WriteHeader(400)
		write.Write(dat)
		return
	}
	if len(params.Body) > 140 {
		respBody := errReturn{
			Error: "Chirp is too long",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			write.WriteHeader(500)
			return
		}
		write.Header().Set("Content-Type", "application/json")
		write.WriteHeader(400)
		write.Write(dat)
		return
	}

	cleaned_body := removeProfanity(params.Body)

	respBody := validReturn{
		CleanedBody: cleaned_body,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		write.WriteHeader(500)
		return
	}
	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(200)
	write.Write(dat)
}

func removeProfanity(chirp string) string {
	profane_words := []string{"kerfuffle", "sharbert", "fornax"}
	split_str := strings.Split(chirp, " ")
	for idx, word := range split_str {
		if slices.Contains(profane_words, strings.ToLower(word)) {
			split_str[idx] = "****"
		}
	}
	cleaned_string := strings.Join(split_str, " ")
	return cleaned_string
}

func (cfg *apiConfig) createNewUser(write http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	params := newUser{}
	err := decoder.Decode(&params)
	if err != nil {
		respBody := errReturn{
			Error: "Something went wrong creating a user",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			write.WriteHeader(500)
			return
		}
		write.Header().Set("Content-Type", "application/json")
		write.WriteHeader(400)
		write.Write(dat)
		return
	}

	newUser, err := cfg.dbQueries.CreateUser(request.Context(), params.Email)
	if err != nil {
		respBody := errReturn{
			Error: "Something went wrong creating a user",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			write.WriteHeader(500)
			return
		}
		write.Header().Set("Content-Type", "application/json")
		write.WriteHeader(400)
		write.Write(dat)
		return
	}

	responseUser := User{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	}

	dat, err := json.Marshal(responseUser)

	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(201)
	write.Write(dat)
}
