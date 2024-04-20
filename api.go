package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
	log.Println("Successful response")
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
	type response struct {
		Feed       database.JSONFeed   `json:"feed"`
		FeedFollow database.FeedFollow `json:"feed_follow"`
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

	feedFollow, err := self.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		FeedID:    feed.ID,
		UserID:    user.ID,
	})

	respondWithJSON(w, http.StatusOK, response{feed.Json(), feedFollow})
}

func (self *apiConfig) getAllFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := self.DB.GetAllFeeds(r.Context())
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "Unable to get the feeds")
		return
	}

	jsonFeeds := make([]database.JSONFeed, len(feeds))
	for i := 0; i < len(feeds); i++ {
		jsonFeeds[i] = feeds[i].Json()
	}

	respondWithJSON(w, http.StatusOK, jsonFeeds)
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

func (self *apiConfig) getUserFeedFollows(w http.ResponseWriter, r *http.Request, user database.User) {
	feefFollows, err := self.DB.GetUserFeedFollows(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, feefFollows)
}

func (self *apiConfig) deleteFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	ffID := r.PathValue("ffID")
	feedFollowID, err := uuid.Parse(ffID)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, err.Error())
	}

	feefFollows, err := self.DB.GetUserFeedFollows(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	isUserFF := false
	for _, ff := range feefFollows {
		if ff.ID == feedFollowID {
			isUserFF = true
			break
		}
	}

	if !isUserFF {
		respondWithError(w, http.StatusUnauthorized, "Deletion not allowed")
		return
	}

	if err = self.DB.DeleteFeedFollow(r.Context(), feedFollowID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create feed follow")
		return
	}
	respondWithJSON(w, http.StatusOK, "")
}

func (self *apiConfig) getPostsForUser(w http.ResponseWriter, r *http.Request, user database.User) {
	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 5
	}

	posts, err := self.DB.GetPostByUser(r.Context(), database.GetPostByUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get posts for user")
	}
	respondWithJSON(w, http.StatusOK, posts)
}
