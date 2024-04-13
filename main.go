package main

import (
	"blagg/internal/database"
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	const filepathRoot = "."
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	godotenv.Load()
	port := os.Getenv("PORT")
	dbURL := os.Getenv("SQL_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Unable to connect to database.")
	}
	config := apiConfig{database.New(db)}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/ok", getHealthCheck)
	mux.HandleFunc("GET /v1/err", getErrorCheck)
	mux.HandleFunc("POST /v1/users", config.postCreateUser)
	mux.HandleFunc("GET /v1/users", config.middlewareAuth(config.getCurrentUser))
	mux.HandleFunc("POST /v1/feeds", config.middlewareAuth(config.postCreateFeed))
	mux.HandleFunc("GET /v1/feeds", config.getAllFeeds)
	mux.HandleFunc("POST /v1/feed_follows", config.middlewareAuth(config.postCreateFeedFollow))

	corsMux := middlewareCors(mux)
	server := &http.Server{Addr: ":" + port, Handler: corsMux}
	log.Fatal(server.ListenAndServe())
}
