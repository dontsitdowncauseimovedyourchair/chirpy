package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type jsonError struct {
		Error string `json:"error"`
	}

	errorJSON := jsonError{Error: msg}
	marshalled, err := json.Marshal(errorJSON)
	if err != nil {
		log.Printf("flop marshalling error response: %s\n", err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(marshalled)
	if err != nil {
		log.Printf("flopped writing %v into response.\n", marshalled)
	}
}

func respondWithJson(w http.ResponseWriter, code int, payload any) {
	marshalled, err := json.Marshal(payload)
	if err != nil {
		log.Printf("flop marshalling successful response: %s\n", err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(marshalled)
	if err != nil {
		log.Printf("flopped writing %v into response.\n", marshalled)
	}
}
