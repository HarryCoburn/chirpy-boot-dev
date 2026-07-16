package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type chirpPost struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type chirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

func servHealth(w http.ResponseWriter, r *http.Request) {
	respondWith(w, http.StatusOK, contentTypePlain, "OK")
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

func (cfg *apiConfig) createNewUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := newUser{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error in createNewUser request: %s", err), "Something went wrong creating a user.")
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error in createNewUser request: %s", err), "Something went wrong hashing a password for a new user.")
		return
	}

	createdUser, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hash,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error in createNewUser request: %s", err), "Something went wrong with the create user query.")
		return
	}

	responseUser := User{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email:     createdUser.Email,
	}

	respondWithJSON(w, http.StatusCreated, responseUser)
}

func (cfg *apiConfig) chirpHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Could not get bearer token: %s", err), "Unauthorized access")
		return
	}

	tokenUUID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Could not validate bearer token: %s", err), "Unauthorized access")
		return
	}

	// Request validation
	decoder := json.NewDecoder(r.Body)
	params := chirpPost{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error in chirp request: %s", err), "Something went wrong creating a chirp.")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", "Chirp is too long.")
		return
	}

	chirp, err := cfg.dbQueries.PostChirp(r.Context(), database.PostChirpParams{
		Body:   removeProfanity(params.Body),
		UserID: tokenUUID,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error in adding chirp to the database: %s", err), "Something went wrong creating a chirp.")
		return
	}

	respondWithJSON(w, http.StatusCreated, toChirpResponse(chirp))
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	allChirps, err := cfg.dbQueries.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Could not retrieve chirps: %s", err), "Something went wrong retrieving chirps.")
		return
	}

	chirpResponses := []chirpResponse{}
	for _, chirp := range allChirps {
		chirpResponses = append(chirpResponses, toChirpResponse(chirp))
	}
	respondWithJSON(w, http.StatusOK, chirpResponses)
}

func (cfg *apiConfig) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirpIDStr := r.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Could not parse chirpID: %s", err), "Something went wrong retreiving chirp.")
		return
	}

	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Could not find this chirp.", "Could not find this chirp.")
			return
		}
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Query in getChirpHandler failed: %s", err), "Query in getChirpHandler failed.")
		return
	}

	respondWithJSON(w, http.StatusOK, toChirpResponse(chirp))
}

func (cfg *apiConfig) userLogin(w http.ResponseWriter, r *http.Request) {
	// Check request for problems
	decoder := json.NewDecoder(r.Body)
	params := newUser{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Error in user login: %s", err), "Something went wrong logging in.")
		return
	}

	// Check for user through the given email
	dbUser, err := cfg.dbQueries.GetUserFromEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Unknown user: %s", err), "Bad username or password")
		return
	}

	// Check that user's password
	validPassword, err := auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil || !validPassword {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Bad username or password: %s", err), "Bad username or password")
		return
	}

	// All good, create bearer token

	jwt, err := auth.MakeJWT(dbUser.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Unable to make JWT token: %s", err), "Server Error")
		return
	}

	// Create refresh token
	refreshToken := auth.MakeRefreshToken()

	cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		UserID:    dbUser.ID,
	})

	respondWithJSON(w, http.StatusOK, User{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        jwt,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("No authorization token found: %s", err), "Bad username or password")
		return
	}

	tokenDb, err := cfg.dbQueries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "invalid refresh token", "Bad username or password - No rows")
			return
		}
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("database error: %s", err), "Something went wrong")
		return
	}

	if time.Now().UTC().After(tokenDb.ExpiresAt.UTC()) {
		respondWithError(w, http.StatusUnauthorized, "refresh token expired", "Bad username or password - expired")
		return
	}
	if tokenDb.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "refresh token revoked", "Bad username or password - revoked")
		return
	}

	// We're good. Make a new bearer token and update the user.
	jwt, err := auth.MakeJWT(tokenDb.UserID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Unable to make JWT token: %s", err), "Server Error")
		return
	}

	respondWithJSON(w, http.StatusOK, tokenResponse{
		Token: jwt,
	})
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("No authorization token found: %s", err), "Bad username or password")
		return
	}

	tokenDb, err := cfg.dbQueries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "invalid refresh token", "Bad username or password")
			return
		}
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("database error: %s", err), "Something went wrong")
		return
	}

	cfg.dbQueries.RevokeToken(r.Context(), tokenDb.Token)
	respondWith(w, 204, contentTypePlain, "")
}
