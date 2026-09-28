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
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("./")))))
	// serve /healthz
	mux.Handle("/healthz", middlewareLog(getServerStatus))
	// serve /metrics
	mux.Handle("/metrics", middlewareLog(apiCfg.getFileServerHitsMetrics))
	// serve /reset
	mux.Handle("/reset", middlewareLog(apiCfg.resetServerHitsMetrics))

	
	// new server struct
	s := &http.Server {
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(s.ListenAndServe())

}

