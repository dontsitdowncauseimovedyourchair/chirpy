package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type jsonRequest struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	var jreq jsonRequest
	err := decoder.Decode(&jreq)
	if err != nil {
		log.Printf("Flop decoding params: %s\n", err.Error())
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if len(jreq.Body) > 140 {
		log.Printf("Long request: %s\n", jreq.Body)
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	out := struct {
		CleanedBody string `json:"cleaned_body"`
	}{
		CleanedBody: filterBadWords(jreq.Body),
	}
	respondWithJson(w, http.StatusOK, out)
}

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

func filterBadWords(msg string) string {
	words := strings.Split(msg, " ")
	badWords := []string{"KERFUFFLE", "SHARBERT", "FORNAX"}
	for i := range words {
		if slices.Contains(badWords, strings.ToUpper(words[i])) {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}
