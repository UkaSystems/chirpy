package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
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

// func (cfg *apiConfig) resetHitCounter(w http.ResponseWriter, r *http.Request) {
// 	cfg.fileserverHits.Store(0)
// 	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("Hits counter reset to 0\n"))
// }

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

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type userPostReq struct {
		Email string `json:"email"`
	}

	type userPostResp struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	req := userPostReq{}
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), req.Email)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating user: %v", err), http.StatusInternalServerError)
		return
	}

	resp := userPostResp{
		Id:        user.ID,
		CreatedAt: user.CreatedAt.String(),
		UpdatedAt: user.UpdatedAt.String(),
		Email:     user.Email,
	}
	respondWithJSON(w, http.StatusCreated, resp)
}

func (cfg *apiConfig) adminResetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "403 Forbidden", errors.New("Not allowed"))
		return
	}

	err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "500 Internal Server Error", err)
		return
	}
	cfg.fileserverHits.Store(0)
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "All users deleted successfully; hit counter reset to 0"})
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAllChirpsOrderedByCreatedAtAsc(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "500 Internal Server Error", err)
		return
	}

	var sliceResp []ChirpResponse
	for _, chirp := range chirps {
		sliceResp = append(sliceResp, ChirpResponseFromDB(&chirp))
	}

	respondWithJSON(w, http.StatusOK, sliceResp)
}

func (cfg *apiConfig) getChirpByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the chirp ID from the URL path
	idStr := r.URL.Path[len("/api/chirps/"):]
	chirpID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "400 Bad Request", fmt.Errorf("invalid chirp ID: %v", err))
		return
	}

	chirp, err := cfg.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	resp := ChirpResponseFromDB(&chirp)
	respondWithJSON(w, http.StatusOK, resp)
}
