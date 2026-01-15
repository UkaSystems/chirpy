package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"sync/atomic"

	"os"

	"github.com/UkaSystems/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println(fmt.Errorf("unable to connect to database: %v", err))
		return
	}
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		db:             database.New(db),
		platform:       os.Getenv("PLATFORM"),
	}

	var serveMux = http.NewServeMux()

	// serve static files from the current directory under /app/:
	var fileServer = http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServer))

	// readiness probe endpoint:
	serveMux.Handle("GET /api/healthz", http.HandlerFunc(readinessHandler))
	serveMux.Handle("GET /api/metrics", http.HandlerFunc(apiCfg.logRequestsNum))
	serveMux.Handle("GET /admin/metrics", http.HandlerFunc(apiCfg.adminMetricsHandler))

	serveMux.Handle("POST /api/chirps", http.HandlerFunc(apiCfg.handlerChirpsValidate))
	serveMux.Handle("POST /api/users", http.HandlerFunc(apiCfg.createUserHandler))
	serveMux.Handle("POST /admin/reset", http.HandlerFunc(apiCfg.adminResetHandler))
	serveMux.Handle("GET /api/chirps", http.HandlerFunc(apiCfg.getChirpsHandler))
	serveMux.Handle("GET /api/chirps/{id}", http.HandlerFunc(apiCfg.getChirpByIDHandler))
	var server = &http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	fmt.Printf("Started server at http://localhost%v\n", server.Addr)
	server.ListenAndServe()
}
