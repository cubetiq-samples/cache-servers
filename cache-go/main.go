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
var writeCacheCh = make(chan map[string]string, 100)

func getHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Key is required"})
		return
	}

	log.Printf("GET key %s", key)
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
	if cache[key] != "" {
		log.Printf("Key %s already exists, overwriting", key)
	}
	cache[key] = value
	lock.Unlock()

	// Notify the persister that there are changes
	writeCacheCh <- cache

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Value stored"})
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	log.Printf("DELETE key %s", key)

	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Key is required"})
		return
	}

	lock.Lock()
	delete(cache, key)
	lock.Unlock()

	// Notify the persister that there are changes
	writeCacheCh <- cache

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Key deleted"})
}

func getKeysHandler(w http.ResponseWriter, r *http.Request) {
	keys := make([]string, 0, len(cache))
	for k := range cache {
		keys = append(keys, k)
	}

	// Get the total size of the cache that store in memory (in bytes)
	cacheSize := 0
	for _, v := range cache {
		cacheSize += len(v)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"keys": keys, "size": cacheSize})
}

func persistCacheWorker() {
	for {
		select {
		case c := <-writeCacheCh:
			persistCache(c)
		}
	}
}

func persistCache(c map[string]string) {
	bytes, err := json.Marshal(c)
	if err != nil {
		fmt.Println("Error marshaling cache:", err)
		return
	}
	err = ioutil.WriteFile("cache.json", bytes, 0644)
	if err != nil {
		fmt.Println("Error writing cache to disk:", err)
	}

	fmt.Println("Cache persisted")
}

func loadCache() error {
	bytes, err := ioutil.ReadFile("cache.json")
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &cache)
}

type ExtendedRequest struct {
	*http.Request
}

type ExtendedResponseWriter struct {
	http.ResponseWriter
}

func (r *ExtendedRequest) Methods(methods ...string) bool {
	for _, method := range methods {
		if r.Method == method {
			return true
		}
	}
	return false
}

func (w *ExtendedResponseWriter) MethodNotAllowedResponse() {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
}

func routerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "GET" {
			next.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path

		switch path {
		case "/get":
			if !(&ExtendedRequest{r}).Methods("GET") {
				(&ExtendedResponseWriter{w}).MethodNotAllowedResponse()
				return
			}
		case "/set":
			if !(&ExtendedRequest{r}).Methods("POST") {
				(&ExtendedResponseWriter{w}).MethodNotAllowedResponse()
				return
			}
		case "/delete":
			if !(&ExtendedRequest{r}).Methods("DELETE") {
				(&ExtendedResponseWriter{w}).MethodNotAllowedResponse()
				return
			}
		default:
			if !(&ExtendedRequest{r}).Methods("GET") {
				(&ExtendedResponseWriter{w}).MethodNotAllowedResponse()
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func printRoutes() {
	fmt.Println("Available routes:")
	fmt.Println("\n------------------------------------------------------------------------")
	fmt.Println("| Method | Route                           | Description                 |")
	fmt.Println("|--------|---------------------------------|-----------------------------|")
	fmt.Println("| GET    | /keys                           | Retrieve all keys of cache  |")
	fmt.Println("| GET    | /get?key={key}                  | Retrieve value for given key|")
	fmt.Println("| POST   | /set?key={key}&value={value}    | Add new key-value to cache  |")
	fmt.Println("| DELETE | /delete?key={key}               | Delete key from cache       |")
	fmt.Println("--------------------------------------------------------------------------")
}

func NewRouter() *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("/keys", getKeysHandler)
	router.HandleFunc("/get", getHandler)
	router.HandleFunc("/set", setHandler)
	router.HandleFunc("/delete", deleteHandler)

	printRoutes()
	return router
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

	go persistCacheWorker()

	router := NewRouter()
	fmt.Println("Starting Redis-like server listening on", addr)
	err = http.ListenAndServe(addr, routerMiddleware(router))
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
