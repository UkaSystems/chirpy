package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
	}

	var serveMux = http.NewServeMux()

	// serve static files from the current directory under /app/:
	var fileServer = http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServer))

	// readiness probe endpoint:
	serveMux.Handle("GET /healthz", http.HandlerFunc(readinessHandler))
	serveMux.Handle("GET /metrics", http.HandlerFunc(apiCfg.logRequestsNum))
	serveMux.Handle("POST /reset", http.HandlerFunc(apiCfg.resetHitCounter))

	var server = &http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	fmt.Printf("Started server at http://localhost%v\n", server.Addr)
	server.ListenAndServe()
}
