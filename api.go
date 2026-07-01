package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/HarryCoburn/chirpy-boot-dev/internal/database"
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

type chirpPost struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type chirpPostValid struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func servHealth(write http.ResponseWriter, request *http.Request) {
	write.Header().Set("Content-Type", "text/plain; charset=utf-8")
	write.WriteHeader(http.StatusOK)
	write.Write([]byte("OK"))
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

func (cfg *apiConfig) chirpHandler(write http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	params := chirpPost{}
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

	var postParams database.PostChirpParams
	postParams.Body = removeProfanity(params.Body)
	postParams.UserID = params.UserID

	postChirpReturn, err := cfg.dbQueries.PostChirp(request.Context(), postParams)

	chirpResponse := chirpPostValid{}
	chirpResponse.ID = postChirpReturn.ID
	chirpResponse.CreatedAt = postChirpReturn.CreatedAt
	chirpResponse.UpdatedAt = postChirpReturn.UpdatedAt
	chirpResponse.Body = postChirpReturn.Body
	chirpResponse.UserID = postChirpReturn.UserID

	dat, err := json.Marshal(chirpResponse)
	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(201)
	write.Write(dat)

}
