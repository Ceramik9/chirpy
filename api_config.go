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
  "github.com/Ceramik9/chirpy/internal/auth"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
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
	type requestUser struct {
		Email          string `json:"email"`
		Password       string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	user := requestUser {}
	err := decoder.Decode(&user)
	if err != nil {
		resBody := []byte(`{"error": "error reading request body"}`)
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(400)
    w.Write(resBody)
		return
	}

	// create user
	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(500)
		w.Write([]byte(`{"body": "error hashing password"}`))
		log.Printf("error hashing password: %v", err)
	}
	userParams := database.CreateUserParams {
		Email: user.Email,
		HashedPassword: hashedPassword,
	}
	newUser, err := cfg.db.CreateUser(r.Context(), userParams)
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
	

	// create chirp request
	type userRequest struct {
		Body   string    `json:"body"`
	}

	// decode request
	decoder := json.NewDecoder(r.Body)
	request := userRequest {}
	err := decoder.Decode(&request)

	// check for errors
	if err != nil {
		resBody := []byte(`{"body": "Error: failed to decode user request"}`)
		log.Print("Error: failed to decode user request")
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(400)
    w.Write(resBody)
		return
	}

	// get user token
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(401)
		w.Write([]byte(`{"body": "error getting user token"}`))
		log.Printf("error getting user token: %v", err)
		return
	}
	
	// validate user
	validatedUserID, err :=auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(401)
		w.Write([]byte(`{"body": "authorization failed"}`))
		log.Printf("authorization failed: %v", err)
		return
	}

	// check chirp length
	if len(request.Body) > 140 {
		resBody := []byte(`{"body": Error: chirp is too long"}`)
		log.Printf("Error: chirp is too long")
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(400)
    w.Write(resBody)
		return
	}
	
	// validate chirp
	validatedChirpBody := profanityFilter(request.Body)

	// add chirp to database
	userID := uuid.NullUUID {
		UUID:  validatedUserID,
		Valid: true,
	}
	chirpParams := database.CreateChirpParams {
		Body:   validatedChirpBody,
		UserID: userID,
	}
	newChirp, err := cfg.db.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(500)
		w.Write([]byte(`{"body": "error adding chirp to database"}`))
		log.Printf("test: %v", err)
		return
	}
	
	//create response
	type chirpResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}
	response := chirpResponse {
		ID:        newChirp.ID,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
		Body:      newChirp.Body,
		UserID:    newChirp.UserID.UUID,
	}
	data, err := json.Marshal(response)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(201)
		w.Write([]byte(`{"body": "error adding chirp to database"}`))
		log.Printf("error marshaling response: %v", err)
		return
	}
	
	// success response
	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(201)
  w.Write(data)
}

func (cfg *apiConfig) getAllChirps(w http.ResponseWriter, r *http.Request) {
	// create container struct for chirp
	type chirp struct {
		ID        uuid.UUID     `json:"id"`
		CreatedAt time.Time     `json:"created_at"`
		UpdatedAt time.Time     `json:"updated_at"`
		Body      string        `json:"body"`
		UserID    uuid.NullUUID `json:"user_id"`
	}
	
	// get all chirps from database
	allChirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(500)
		w.Write([]byte(`{"body": "error getting chirps from database"}`))
		log.Printf("error getting chirps from database: %v", err)
		return
	}
	
	// convert chirp keys
	chirpsSlice := make([]chirp, 0, len(allChirps))
	for i := 0; i < len(allChirps); i++ {
		chirpsSlice = append(chirpsSlice, chirp {
			ID: allChirps[i].ID,
			CreatedAt: allChirps[i].CreatedAt,
			UpdatedAt: allChirps[i].UpdatedAt,
			Body: allChirps[i].Body,
			UserID: allChirps[i].UserID,
		})
	}

	// marshal response body
	data, err := json.Marshal(chirpsSlice)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(500)
		w.Write([]byte(`{"body": "error marshaling response body"}`))
		log.Printf("error marshaling response body: %v", err)
		return
	}
	
	// success response
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
		w.Write(data)
}

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {

	// parse chirp id
	chirpID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(400)
		w.Write([]byte(`{"body": "error parsing user id"}`))
		log.Printf("error parsing user id: %v", err)
	}
	
	// get the cirp with matching ID from db
	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(404)
		w.Write([]byte(`{"body": "error getting chirp from database"}`))
		log.Printf("error getting chifrp from database: %v", err)
	}
	type response struct {
		ID        uuid.UUID     `json:"id"`
		CreatedAt time.Time     `json:"created_at"`
		UpdatedAt time.Time     `json:"updated_at"`
		Body      string        `json:"body"`
		UserID    uuid.NullUUID `json:"user_id"`
	}
	
	res := response {
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
	data, err := json.Marshal(res)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(500)
		w.Write([]byte(`{"body": "error marshalling response"}`))
		log.Printf("error marshalling response : %v", err)
	}

	// success response
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
		w.Write(data)
}

func (cfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {

	// decode user request
	type userRequest struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	request := userRequest {}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(401)
		w.Write([]byte(`{"body": "error decoding user request"}`))
		log.Printf("error decoding user request : %v", err)
		return
	}

	// get uesr's hashed password
	dbUser, err := cfg.db.GetUser(r.Context(), request.Email)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(401)
		w.Write([]byte(`{"body": "error getting user"}`))
		log.Printf("error getting user : %v", err)
		return
	}

	//authenticate user
	match, err := auth.CheckPasswordHash(request.Password, dbUser.HashedPassword)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(401)
		w.Write([]byte(`{"body": "error checking password"}`))
		log.Printf("error checking password : %v", err)
		return
	}
	if !match {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(401)
		w.Write([]byte(`{"body": "incorrect password"}`))
		log.Printf("incorrect password : %v", nil)
		return
	}
	// success response
	type responseBody struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}

	// create session token
	var expirietionTime time.Duration
	if request.ExpiresInSeconds < 1 || request.ExpiresInSeconds > 3600 {
		expirietionTime = time.Duration(3600)*time.Second
	} else {
		expirietionTime = time.Duration(request.ExpiresInSeconds) * time.Second
	}

	sessionToken, err := auth.MakeJWT(dbUser.ID, cfg.secret, expirietionTime)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(500)
		w.Write([]byte(`{"body": "error creating session token"}`))
		log.Printf("error creating session token: %v", err)
		return
	}
	response := responseBody {
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		Token:     sessionToken,
	}
	data, err := json.Marshal(response)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(500)
		w.Write([]byte(`{"body": "error marshaling json"}`))
		log.Printf("error marshaling json : %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(200)
	w.Write(data)
}








