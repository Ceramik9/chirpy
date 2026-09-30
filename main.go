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

	// load env
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	// load database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading database: %v", err)
		os.Exit(1)
	}
	apiCfg.db = database.New(db)

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

	
	// new server struct
	s := &http.Server {
		Addr:    ":8080",
		Handler: mux,
	}
	// start listening for requets
	log.Fatal(s.ListenAndServe())

}

