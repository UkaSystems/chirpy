package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

// Boot.dev approach to the JSON API handlers to validate chirp length:
func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

var profanityList = []string{
	"kerfuffle",
	"sharbert",
	"fornax",
}

func CleanChirp(body string) string {

	split := strings.Split(body, " ")
	for i, word := range split {
		lowerWord := strings.ToLower(word)
		if slices.Contains(profanityList, lowerWord) {
			split[i] = strings.Repeat("*", 4)
		}
	}

	return strings.Join(split, " ")
}

func handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		Valid       bool   `json:"valid"`
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		Valid:       true,
		CleanedBody: CleanChirp(params.Body),
	})
}
