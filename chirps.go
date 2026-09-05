package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/dontsitdowncauseimovedyourchair/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
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

func (cfg *apiConfig) handleChirpsPost(w http.ResponseWriter, r *http.Request) {
	type ChirpReq struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	var chirp ChirpReq
	err := decoder.Decode(&chirp)
	if err != nil {
		log.Printf("Flop decoding chirp: %s\n", err.Error())
		respondWithError(w, http.StatusBadRequest, "Invalid request payload or user ID")
		return
	}

	if chirp.UserID == uuid.Nil {
		respondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	if len(chirp.Body) > 140 {
		log.Printf("Long request: %s\n", chirp.Body)
		respondWithError(w, http.StatusBadRequest, "Chirp is too long, max 140 characters")
		return
	}

	chirp.Body = filterBadWords(chirp.Body)

	dbChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code.Name() == "foreign_key_violation" {
			respondWithError(w, http.StatusBadRequest, "Invalid user ID: user does not exist")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Flop creating chirp :(. Sorry")
		return
	}

	respondWithJson(w, http.StatusCreated, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}
