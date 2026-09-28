package main

import (
	"net/http"
	"log"
)

func main() {
	

	// Create API Config
	var apiCfg apiConfig


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

	
	// new server struct
	s := &http.Server {
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(s.ListenAndServe())

}

