package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/ok", getHealthCheck)
	mux.HandleFunc("GET /v1/err", getErrorCheck)

	corsMux := middlewareCors(mux)
	server := &http.Server{Addr: ":" + port, Handler: corsMux}
	log.Fatal(server.ListenAndServe())
}
