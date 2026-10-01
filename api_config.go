package main

import(
	"log"
	"time"
	"sync/atomic"
	"net/http"
	"fmt"
	"github.com/Ceramik9/chirpy/internal/database"
	"encoding/json"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
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
	
	// check environment
	if cfg.platform != "dev" {
		w.WriteHeader(403)
		w.Write([]byte("Forbidden"))
	}

	// write status code (200)
	w.WriteHeader(http.StatusOK)
	// write response body
	message := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())
	w.Write([]byte(message))
}

func (cfg *apiConfig) resetServerHitsMetrics(w http.ResponseWriter, r *http.Request) {
	
	// delete all users from database
	cfg.db.DeleteAllUsers(r.Context())
	//reset fileserverHits
	cfg.fileserverHits.Store(0)
	// set header
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// write status code (200)
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {

	// decode request body
	type requestEmail struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	user := requestEmail {}
	err := decoder.Decode(&user)
	if err != nil {
		resBody := []byte(`{"error": "error reading request body"}`)
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(400)
    w.Write(resBody)
		return
	}

	// create user
	newUser, err := cfg.db.CreateUser(r.Context(), user.Email)
	if err != nil {
		resBody := []byte(`{"error": "error creating new user"}`)
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(500)
    w.Write(resBody)
		return
	}

	// create response
	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	res := response {
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	}

	data, err := json.Marshal(res)
	if err != nil {
		resBody := []byte(`{"error": "error creating response"}`)
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(201)
    w.Write(resBody)
		return
	}
	
	// success response
	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(201)
  w.Write(data)
}


func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	
	// create request struct
	type chirpRequest struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}


}


func errorHandler(description string, err error) ([]byte, error) {

	type errorHolder struct {
		desc         string `json:"description"`
		errorMessage error  `json:"error"`
	}

	newError := errorHolder {
		desc:         description,
		errorMessage: err,
	}

	data, err := json.Marshal(newError)
	if err != nil {
		return nil, fmt.Errorf("failed to create error response: %w", err)
	}
	return data, nil
}


