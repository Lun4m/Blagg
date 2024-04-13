package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"blagg/internal/database"
)

type apiConfig struct {
	DB *database.Queries
}

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

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
		return
	}
	user, err := self.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create user")
		return
	}
	respondWithJSON(w, http.StatusOK, user)
}

func (self *apiConfig) middlewareAuth(next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authStr := r.Header.Get("Authorization")
		if !strings.HasPrefix(authStr, "ApiKey ") {
			respondWithError(w, http.StatusUnauthorized, "API key not provided")
			return
		}
		user, err := self.DB.GetUser(r.Context(), authStr[7:])
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		next(w, r, user)
	}
}

func (self *apiConfig) getCurrentUser(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, http.StatusOK, user)
}

func (self *apiConfig) postCreateFeed(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	feed, err := self.DB.CreateFeed(r.Context(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
		Url:       params.Url,
		UserID:    user.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create feed")
		return
	}
	respondWithJSON(w, http.StatusOK, feed)
}

func (self *apiConfig) getAllFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := self.DB.GetAllFeeds(r.Context())
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Unable to get the feeds")
		return
	}
	respondWithJSON(w, http.StatusOK, feeds)
}

func (self *apiConfig) postCreateFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		FeedID uuid.UUID `json:"feed_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	feedFollow, err := self.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		FeedID:    params.FeedID,
		UserID:    user.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create feed follow")
		return
	}
	respondWithJSON(w, http.StatusOK, feedFollow)
}
