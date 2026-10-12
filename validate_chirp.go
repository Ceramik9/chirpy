package main

import(
	"encoding/json"
	"net/http"
	"strings"
	"log"
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
    w.WriteHeader(400)
		log.Printf("error decoding request body: %v", err)
		return
	}
	if len(request.Body) > 140 {
    w.WriteHeader(400)
    log.Print("chirp is too long")
		return
	}

	// profanity filter
	var resBody = responseBody {
		Message: profanityFilter(request.Body),
	}
	data, err := json.Marshal(resBody)
	if err != nil {
  	w.WriteHeader(400)
		log.Printf("error marshalling response body: %v", err)
		return
	}
	
	// success
	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(200)
  w.Write(data)
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


