package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	servemux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":" + port,
		Handler: servemux,
	}
	log.Fatal(server.ListenAndServe())
}
