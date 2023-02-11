package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

var cache = make(map[string]string)
var lock sync.RWMutex

func getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	lock.RLock()

	value, found := cache[key]
	lock.RUnlock()

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
	lock.Lock()
	cache[key] = value
	lock.Unlock()

	err := persistCache()
	if err != nil {
		log.Println("Error persisting cache:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error persisting cache"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Value stored"})
}

func persistCache() error {
	lock.RLock()
	defer lock.RUnlock()

	bytes, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return ioutil.WriteFile("cache.json", bytes, 0644)
}

func loadCache() error {
	bytes, err := ioutil.ReadFile("cache.json")
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &cache)
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
	err := loadCache()
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("Error loading cache:", err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	host := os.Getenv("HOST")

	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	http.HandleFunc("/cache", cacheHandler)
	fmt.Println("Starting Redis-like server listening on", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
