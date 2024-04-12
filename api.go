package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	data, _ := json.Marshal(errorResponse{msg})
	log.Printf("Responding with ERROR: %s", msg)
	w.WriteHeader(code)
	w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("Error marshalling JSON: %s", err))
		return
	}
	log.Printf("Responding with value: %v", payload)
	w.WriteHeader(code)
	w.Write(data)
}
