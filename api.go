package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/HarryCoburn/chirpy-boot-dev/internal/auth"
	"github.com/HarryCoburn/chirpy-boot-dev/internal/database"
	"github.com/google/uuid"
)

type parameters struct {
	Body string `json:"body"`
}

type errReturn struct {
	Error string `json:"error"`
}

type newUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respBody := errReturn{
			Error: "Something went wrong creating a user",
		}
		dat, _ := json.Marshal(respBody)
		write.Header().Set("Content-Type", "application/json")
		write.WriteHeader(400)
		write.Write(dat)
		return
	}

	params.Password = hash

	newUser, err := cfg.dbQueries.CreateUser(request.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hash,
	})
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

	chirpResponse := chirpPostValid{
		ID:        postChirpReturn.ID,
		CreatedAt: postChirpReturn.CreatedAt,
		UpdatedAt: postChirpReturn.UpdatedAt,
		Body:      postChirpReturn.Body,
		UserID:    postChirpReturn.UserID,
	}

	dat, err := json.Marshal(chirpResponse)
	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(201)
	write.Write(dat)

}

func (cfg *apiConfig) getChirpsHandler(write http.ResponseWriter, request *http.Request) {
	chirpArray := []chirpPostValid{}
	allChirps, err := cfg.dbQueries.GetAllChirps(request.Context())
	if err != nil {
		fmt.Println("Could not retreive all chirps.")
		return
	}
	for _, chirp := range allChirps {
		chirpResponse := chirpPostValid{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		chirpArray = append(chirpArray, chirpResponse)
	}
	dat, err := json.Marshal(chirpArray)
	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(200)
	write.Write(dat)
}

func (cfg *apiConfig) getChirpHandler(write http.ResponseWriter, request *http.Request) {
	chirpIDStr := request.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		log.Fatalf("failed to parse UUID string: %v", err)
	}
	chirp, err := cfg.dbQueries.GetChirp(request.Context(), chirpUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			write.Header().Set("Content-Type", "application/json")
			write.WriteHeader(404)
			fmt.Println("Could not find this chirp")
			return
		}
		fmt.Println("Query in getChirpHandler failed.")
		return

	}

	chirpResponse := chirpPostValid{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	dat, err := json.Marshal(chirpResponse)
	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(200)
	write.Write(dat)
}

func (cfg *apiConfig) userLogin(write http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)
	params := newUser{}
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

	user, err := cfg.dbQueries.GetUserFromEmail(request.Context(), params.Email)
	if err != nil {
		respBody := errReturn{
			Error: "Incorrect email or password",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			write.WriteHeader(500)
			return
		}
		write.Header().Set("Content-Type", "application/json")
		write.WriteHeader(401)
		write.Write(dat)
		return
	}

	did_auth, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || did_auth == false {
		respBody := errReturn{
			Error: "Incorrect email or password",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			write.WriteHeader(500)
			return
		}
		write.Header().Set("Content-Type", "application/json")
		write.WriteHeader(401)
		write.Write(dat)
		return
	}

	responseUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	dat, err := json.Marshal(responseUser)

	write.Header().Set("Content-Type", "application/json")
	write.WriteHeader(200)
	write.Write(dat)

}
