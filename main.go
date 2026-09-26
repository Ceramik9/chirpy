package main

import (
	"net/http"
	"log"
)

func main() {
	
	//new ServeMux
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("./"))))
	mux.HandleFunc("/healthz", getServerStatus)
	
	// new server struct
	s := &http.Server {
		Addr: 	":8080",
		Handler: mux,
	}

	log.Fatal(s.ListenAndServe())

}

