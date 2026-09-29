package main

import(
	"encoding/json"
	"net/http"
	"strings"
)

func validateChirp(w http.ResponseWriter, r *http.Request) {

	// chirp holder
	type chirp struct {
		Body string `json:"body"`
	}

	type responseBody struct {
		Message string `json:"cleaned_body"`
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

	// profanity filter
	var resBody = responseBody {
		Message: profanityFilter(request.Body),
	}
	data, err := json.Marshal(resBody)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
  	w.WriteHeader(400)
		w.Write([]byte(`{"error": "Failed to generate response body"}`))
		return
	}
	
	// success
	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(200)
  w.Write(data)
	return
}


func profanityFilter(chirp string) string {

	// split words
	words := strings.Split(chirp, " ")

	// filter words
	for i := 0 ; i < len(words); i++ {
		if _, exists := profanity[strings.ToLower(words[i])]; exists {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}


var profanity = map[string]struct{} {
	"kerfuffle": {},
	"sharbert": {},
	"fornax": {},
}


