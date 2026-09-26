package main

import(
	"net/http"
)

func getServerStatus(res http.ResponseWriter, req *http.Request) {
	
	// set header
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("OK"))
}
