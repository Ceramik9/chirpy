package main

import(
	"log"
	"time"
	"sync/atomic"
	"net/http"
	"fmt"
	"sort"
	"github.com/Ceramik9/chirpy/internal/database"
	"encoding/json"
	"github.com/google/uuid"
  "github.com/Ceramik9/chirpy/internal/auth"
	"database/sql"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
	polka          string
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
    w.WriteHeader(400)
		log.Printf("error reading request body: %v", err)
		return
	}

	// create user
	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
    w.WriteHeader(500)
		log.Printf("error hashing password: %v", err)
		return
	}
	userParams := database.CreateUserParams {
		Email: user.Email,
		HashedPassword: hashedPassword,
	}
	newUser, err := cfg.db.CreateUser(r.Context(), userParams)
	if err != nil {
    w.WriteHeader(500)
		log.Printf("error creating new user: %v", err)
		return
	}

	// create response
	type response struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	res := response {
		ID:          newUser.ID,
		CreatedAt:   newUser.CreatedAt,
		UpdatedAt:   newUser.UpdatedAt,
		Email:       newUser.Email,
		IsChirpyRed: newUser.IsChirpyRed,
	}

	data, err := json.Marshal(res)
	if err != nil {
    w.WriteHeader(500)
		log.Printf("error marshaling response: %v", err)
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
		log.Printf("error decoding request: %v", err)
    w.WriteHeader(400)
		return
	}

	// get user token
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
    w.WriteHeader(401)
		log.Printf("error getting user token: %v", err)
		return
	}
	
	// validate user
	validatedUserID, err :=auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
    w.WriteHeader(401)
		log.Printf("authorization failed: %v", err)
		return
	}

	// check chirp length
	if len(request.Body) > 140 {
    w.WriteHeader(400)
		log.Print("error, chirp is too long")
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
    w.WriteHeader(500)
		log.Printf("error adding chirp to database: %v", err)
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
    w.WriteHeader(500)
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
	dbChirps := []database.Chirp {}
	var err error
	userQuerry := r.URL.Query().Get("author_id")
	userID, _ := uuid.Parse(userQuerry)
	nullUserID := uuid.NullUUID {
		UUID:  userID,
		Valid: true,
	}
	if userQuerry != "" {
		dbChirps, err = cfg.db.GetAllChirpsForUser(r.Context(), nullUserID)
	} else {
		dbChirps, err = cfg.db.GetAllChirps(r.Context())
	}
	if err != nil {
    w.WriteHeader(500)
		log.Printf("error getting chirps from database: %v", err)
		return
	}
	
	// convert chirp keys
	chirpsSlice := make([]chirp, 0, len(dbChirps))
	for i := 0; i < len(dbChirps); i++ {
		chirpsSlice = append(chirpsSlice, chirp {
			ID: dbChirps[i].ID,
			CreatedAt: dbChirps[i].CreatedAt,
			UpdatedAt: dbChirps[i].UpdatedAt,
			Body: dbChirps[i].Body,
			UserID: dbChirps[i].UserID,
		})
	}

	// sort chirps
	sortQuerry := r.URL.Query().Get("sort")
	if sortQuerry == "desc" {
		sort.Slice(chirpsSlice, func(i, j int) bool {
			return chirpsSlice[i].CreatedAt.After(chirpsSlice[j].CreatedAt)
		})
	}

	// marshal response body
	data, err := json.Marshal(chirpsSlice)
	if err != nil {
    w.WriteHeader(500)
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
    w.WriteHeader(400)
		log.Printf("error parsing user id: %v", err)
	}
	
	// get the cirp with matching ID from db
	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
    w.WriteHeader(404)
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
    w.WriteHeader(500)
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
	}
	request := userRequest {}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
    w.WriteHeader(401)
		log.Printf("error decoding user request : %v", err)
		return
	}

	// get uesr's hashed password
	dbUser, err := cfg.db.GetUser(r.Context(), request.Email)
	if err != nil {
    w.WriteHeader(401)
		log.Printf("error getting user : %v", err)
		return
	}

	//authenticate user
	match, err := auth.CheckPasswordHash(request.Password, dbUser.HashedPassword)
	if err != nil {
    w.WriteHeader(401)
		log.Printf("error checking password : %v", err)
		return
	}
	if !match {
    w.WriteHeader(401)
		log.Printf("incorrect password : %v", nil)
		return
	}
	// success response
	type responseBody struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}

	// create access token
	expirationTime := time.Duration(1*time.Hour)
	token, err := auth.MakeJWT(dbUser.ID, cfg.secret, expirationTime)
	if err != nil {
    w.WriteHeader(500)
		log.Printf("error creating session token: %v", err)
		return
	}

	 // create refresh token and add to database
	 refreshToken := auth.MakeRefreshToken()
	 refreshTokenExpiration := time.Now().Add(60 * 24 * time.Hour)
	 refreshTokenParams := database.CreateRefreshTokenParams {
		Token:     refreshToken,
		UserID:    dbUser.ID,
		ExpiresAt: refreshTokenExpiration,
	 }

	 err = cfg.db.CreateRefreshToken(r.Context(), refreshTokenParams)
	 if err != nil {
    w.WriteHeader(500)
		log.Printf("error adding refresh_token to database: %v", err)
		return
	 }

	response := responseBody {
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  dbUser.IsChirpyRed,
	}
	data, err := json.Marshal(response)
	if err != nil {
    w.WriteHeader(500)
		log.Printf("error marshaling json : %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) refreshToken(w http.ResponseWriter, r *http.Request) {
	
	// get user token from request header
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
    w.WriteHeader(401)
		log.Printf("error, invalid token : %v", err)
		return
	}
	
	// get token from database
	dbToken, err := cfg.db.GetRefreshToken(r.Context(), token)
	
	// check if the token exist in database or any other error occured
	if err != nil {
    w.WriteHeader(401)
		log.Printf("error getting token from databse : %v", err)
		return
	}

	// check if the token is valid
	if time.Now().After(dbToken.ExpiresAt) {
    w.WriteHeader(401)
		log.Printf("error, token expired : %v", err)
		return
	} else if dbToken.RevokedAt.Valid {
    w.WriteHeader(401)
		log.Printf("error, token revoked : %v", err)
		return
	}
	
	// create new access token
	expirationTime := time.Duration(1*time.Hour)
	newToken, err := auth.MakeJWT(dbToken.UserID, cfg.secret, expirationTime)
	if err != nil {
    w.WriteHeader(500)
		log.Printf("error creating new access token : %v", err)
		return
	}
	
	// respond with new token
	type responseBody struct {
		Token string `json:"token"`
	}
	response := responseBody {
		Token: newToken,
	}
	res, err := json.Marshal(response)
	if err != nil {
    w.WriteHeader(500)
		log.Printf("error marshalling response body : %v", err)
		return
	}
	// success
	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(200)
	w.Write(res)
}

func (cfg *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {

	// get user token from request header
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
    w.WriteHeader(401)
		log.Printf("error, invalid token : %v", err)
		return
	}
	
	revokeTime := sql.NullTime {
		Time:  time.Now(),
		Valid: true,
	}
	revokeParams := database.RevokeRefreshTokenParams {
		RevokedAt: revokeTime,
		Token:     token,
	}

		err = cfg.db.RevokeRefreshToken(r.Context(), revokeParams)
		if err != nil {
    w.WriteHeader(401)
		log.Printf("error, failed to revoke refresh token: %v", err)
		return
		}
    w.WriteHeader(204)
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {

	// get user token from request header
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
    w.WriteHeader(401)
		log.Printf("error, invalid token : %v", err)
		return
	}

	// authorize
	userID, err := auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
    w.WriteHeader(401)
		log.Printf("error, invalid token : %v", err)
		return
	}

	// decode request body
	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	request := requestBody {}
	err = decoder.Decode(&request)
	if err != nil {
    w.WriteHeader(401)
		log.Printf("error decoding request body: %v", err)
		return
	}
	
	// update user
	hashedPassword, err := auth.HashPassword(request.Password)
	if err != nil {
    w.WriteHeader(401)
		log.Printf("error hashing password: %v", err)
		return
	}

	updateParams := database.UpdateUserParams {
		Email:          request.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	}

	err = cfg.db.UpdateUser(r.Context(), updateParams)
	if err != nil {
    w.WriteHeader(401)
		log.Printf("error updating user: %v", err)
		return
		}

	// create response
	type responseBody struct {
		ID             uuid.UUID `json:"id"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
		Email          string    `json:"email"`
		IsChirpyRed    bool      `json:"is_chirpy_red`
	}

	updatedUser, err := cfg.db.GetUser(r.Context(), request.Email)
	if err != nil {
    w.WriteHeader(500)
		log.Printf("error creating response body: %v", err)
		return
	}

	response := responseBody {
		ID:          updatedUser.ID,
		CreatedAt:   updatedUser.CreatedAt,
		UpdatedAt:   updatedUser.UpdatedAt,
		Email:       updatedUser.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
	}

	data, err := json.Marshal(response)
	if err != nil {
    w.WriteHeader(500)
		log.Printf("error marshalling response body: %v", err)
		return
	}

	// succcess
	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {

// get user token from request header
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
    w.WriteHeader(401)
		log.Printf("error, invalid token : %v", err)
		return
	}

	// authenticate
	userID, err := auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
    w.WriteHeader(401)
		log.Printf("error, invalid token : %v", err)
		return
	}
	
	// get chirp id
	chirpID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
    w.WriteHeader(400)
		log.Printf("error parsing user id: %v", err)
		return
	}

	// authorize
	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
    w.WriteHeader(404)
		log.Printf("error getting chirp from database: %v", err)
		return
	}
	
	if userID != dbChirp.UserID.UUID {
    w.WriteHeader(403)
		log.Printf("error, user is aunauthorized: %v", err)
		return
	}
	
	// delete chirp
	user := uuid.NullUUID {
		UUID:  userID,
		Valid: true,
	}
	
	deleteParamas := database.DeleteChirpParams {
		ID:     chirpID,
		UserID: user,
	}

	err = cfg.db.DeleteChirp(r.Context(), deleteParamas)
	if err != nil {
    w.WriteHeader(400)
		log.Printf("error parsing user id: %v", err)
		return
	}

	// success
  w.WriteHeader(204)
}

func (cfg *apiConfig) updateUserRed(w http.ResponseWriter, r *http.Request) {

	type data struct {
		UserID string `json:"user_id"`
	}

	type requestBody struct {
		Event string `json:"event"`
		Data  data   `json:"data"`
	}
	
	//authorize
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		w.WriteHeader(401)
		log.Printf("error, getting Authorization header: %v", err)
		return
	}

	if apiKey != cfg.polka {
		w.WriteHeader(401)
		log.Printf("unauthorized: %v", err)
		return
	}

	//decode request body
	decoder := json.NewDecoder(r.Body)
	request := requestBody {}
	err = decoder.Decode(&request)
	if err != nil {
    w.WriteHeader(400)
		log.Printf("error, failed to decode request body: %v", err)
		return
	}
	
	// check event type
	if request.Event != "user.upgraded" {
		w.WriteHeader(204)
		log.Printf("not update red request: %v", err)
		return
	}
	parsedUserID, _ := uuid.Parse(request.Data.UserID)
	err = cfg.db.UpdateRed(r.Context(), parsedUserID)
	if err != nil {
		w.WriteHeader(404)
		log.Printf("error updating user's red: %v", err)
	}

	// success
	w.WriteHeader(204)
	w.Write([]byte(`{"body": ""}`))

}


