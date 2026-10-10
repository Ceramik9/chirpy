package main

import (
	"fmt"
	"net/http"
	"log"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"os"
	"database/sql"
	"github.com/Ceramik9/chirpy/internal/database"
)

func main() {

	// Create API Config
	var apiCfg apiConfig

	// load .env
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")

	// load database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading database: %v", err)
		os.Exit(1)
	}
	apiCfg.db = database.New(db)

	//load environment
	apiCfg.platform = platform

	// load secret
	apiCfg.secret = secret
 

	//new ServeMux
 	mux := http.NewServeMux()
	// serve /app
	mux.Handle("GET /app", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("./")))))
	// serve /app/assets
	mux.Handle("GET /app/assets/", http.StripPrefix("/app/assets", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("./assets/")))))
	// serve /api/healthz
	mux.Handle("GET /api/healthz", middlewareLog(getServerStatus))
	// serve /api/metrics
	mux.Handle("GET /admin/metrics", middlewareLog(apiCfg.getFileServerHitsMetrics))
	// serve /api/reset
	mux.Handle("POST /admin/reset", middlewareLog(apiCfg.resetServerHitsMetrics))
	// serve /api/validate_chirp
	mux.Handle("POST /api/validate_chirp", middlewareLog(validateChirp))
	// serve /api/users
	mux.Handle("POST /api/users", middlewareLog(apiCfg.createUser))
	// serve /api/chirps
	mux.Handle("POST /api/chirps", middlewareLog(apiCfg.createChirp))
	// serve /api/chirps
	mux.Handle("GET /api/chirps", middlewareLog(apiCfg.getAllChirps))
	// serve /api/chirps/{id}
	mux.Handle("GET /api/chirps/{id}", middlewareLog(apiCfg.getChirp))
	// serve /api/login
	mux.Handle("POST /api/login", middlewareLog(apiCfg.loginUser))
	// serve /api/refresh
	mux.Handle("POST /api/refresh", middlewareLog(apiCfg.refreshToken))
	// serve /api/revoke
	mux.Handle("POST /api/revoke", middlewareLog(apiCfg.revokeToken))
	// serve /api/users
	mux.Handle("PUT /api/users", middlewareLog(apiCfg.updateUser))

	
	// new server struct
	s := &http.Server {
		Addr:    ":8080",
		Handler: mux,
	}

	// start listening server
	log.Fatal(s.ListenAndServe())

}

