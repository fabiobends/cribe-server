package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(map[string]string{"message": "ok"})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "not found"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusNotFound)

	err := json.NewEncoder(w).Encode(map[string]string{"message": "not found"})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/status", statusHandler)
	mux.HandleFunc("/", notFoundHandler)

	fmt.Println("Server is running on port 8080")

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
	}
}
