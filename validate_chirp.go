package main

import(
	"encoding/json"
	"net/http"
)

func validateChirp(w http.ResponseWriter, r *http.Request) {

	// chirp holder
	type chirp struct {
		Body string `json:"body"`
	}

	// decode json
	decoder := json.NewDecoder(r.Body)
	request := chirp {}
	err := decoder.Decode(&request)
	
	// check for errors
	if err != nil {
		resBody := []byte(`{"error": "Something went wrong"}`)
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(400)
    w.Write(resBody)
		return
	}
	if len(request.Body) > 140 {
		resBody := []byte(`{"error": "Chirp is too long"}`)
		w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(400)
    w.Write(resBody)
		return
	}

	// success
	resBody := []byte(`{"valid": true}`)
	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(200)
  w.Write(resBody)
	return
}

