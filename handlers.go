package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) logRequestsNum(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Hits: %v\n", cfg.fileserverHits.Load())
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hits: %v\n", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) adminMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	html := fmt.Sprintf(
		`<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
		</html>`, cfg.fileserverHits.Load())
	w.Write([]byte(html))
}

func (cfg *apiConfig) resetHitCounter(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits counter reset to 0\n"))
}

func (cfg *apiConfig) validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}

	type chirpValidationResponse struct {
		Error string `json:"error"`
		Valid bool   `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	c := chirp{}
	validity := false
	errResponse := ""
	statusCode := http.StatusOK

	err := decoder.Decode(&c)
	if err != nil {
		errResponse = fmt.Sprintf("Something went wrong: %v", err)
		statusCode = http.StatusBadRequest
	}

	if len(c.Body) <= 140 {
		validity = true
	} else {
		validity = false
		errResponse = "Chirp is too long"
		statusCode = http.StatusBadRequest
	}

	chirpResponse := chirpValidationResponse{
		Valid: validity,
		Error: errResponse,
	}

	dat, err := json.Marshal(chirpResponse)
	if err != nil {
		fmt.Printf("Error marshaling response: %v", err)
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(dat)
}
