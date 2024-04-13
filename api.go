package main

import (
	"blagg/internal/database"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type apiConfig struct {
	DB *database.Queries
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	data, _ := json.Marshal(errorResponse{msg})
	log.Printf("Error: %s", msg)
	w.WriteHeader(code)
	w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	log.Printf("Success: %v", payload)
	w.WriteHeader(code)
	w.Write(data)
}

func (self *apiConfig) postCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name string `json:"name"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, 500, err.Error())
	}
	user, err := self.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create user")
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (self *apiConfig) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	authStr := r.Header.Get("Authorization")
	if !strings.HasPrefix(authStr, "ApiKey ") {
		respondWithError(w, http.StatusUnauthorized, "Wrong API key")
		return
	}

	user, err := self.DB.GetUser(r.Context(), authStr[7:])
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Wrong API key")
		return
	}
	respondWithJSON(w, http.StatusOK, user)
}
