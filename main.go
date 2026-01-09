package main

import (
	"fmt"
	"net/http"
)

func main() {
	var serveMux = http.NewServeMux()

	var fileServer = http.FileServer(http.Dir("."))
	serveMux.Handle("/", fileServer)

	var server = &http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	fmt.Printf("Started server at http://localhost%v\n", server.Addr)
	server.ListenAndServe()
}
