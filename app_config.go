package main

import(
	"log"
	"sync/atomic"
	"net/http"
	"fmt"
	"github.com/Ceramik9/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getFileServerHitsMetrics(w http.ResponseWriter, r *http.Request) {
	
	// set header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// write status code (200)
	w.WriteHeader(http.StatusOK)
	// write response body
	message := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())
	w.Write([]byte(message))
}

func (cfg *apiConfig) resetServerHitsMetrics(w http.ResponseWriter, r *http.Request) {
	
	// set header
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// write status code (200)
	w.WriteHeader(http.StatusOK)
	//reset fileserverHits
	cfg.fileserverHits.Store(0)
}


