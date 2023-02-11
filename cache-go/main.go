package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

var cache = make(map[string]string)

func getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value, found := cache[key]

	if !found {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Key not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"key": key, "value": value})
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("key")
	value := r.FormValue("value")

	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Key is required"})
		return
	}

	log.Printf("SET key %s", key)
	cache[key] = value
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Value stored"})
}

func cacheHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		getHandler(w, r)
	case "POST":
		setHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}

func main() {
	port := os.Getenv("PORT")
	host := os.Getenv("HOST")

	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	http.HandleFunc("/cache", cacheHandler)
	fmt.Println("Starting Redis-like server listening on", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
