package main

import(
	"net/http"
)

func getServerStatus(w http.ResponseWriter, r *http.Request) {
	
	// set header
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// write status code (200)
	w.WriteHeader(http.StatusOK)
	// write response body
	w.Write([]byte("OK"))
}
