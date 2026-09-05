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
