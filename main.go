package main

import (
	"fmt"
	"net/http"
)

func main() {
	var serveMux = http.NewServeMux()

	// serve static files from the current directory under /app/:
	var fileServer = http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serveMux.Handle("/app/", fileServer)

	// readiness probe endpoint:
	serveMux.Handle("/healthz", http.HandlerFunc(readinessHandler))

	var server = &http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	fmt.Printf("Started server at http://localhost%v\n", server.Addr)
	server.ListenAndServe()
}
