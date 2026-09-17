package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/dontsitdowncauseimovedyourchair/chirpy/internal/auth"
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
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "your auth is flop")
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "your auth is flop")
		return
	}

	decoder := json.NewDecoder(r.Body)
	var chirp ChirpReq
	err = decoder.Decode(&chirp)
	if err != nil {
		log.Printf("Flop decoding chirp: %s\n", err.Error())
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
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
		UserID:    userID,
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

func (cfg *apiConfig) handleChirpsGet(w http.ResponseWriter, r *http.Request) {
	var chirps []database.Chirp
	var err error

	rawAuthorID := r.URL.Query().Get("author_id")
	rawSort := r.URL.Query().Get("sort")
	isDesc := "desc" == rawSort

	if len(rawAuthorID) == 0 {
		chirps, err = cfg.db.FetchAllChirps(r.Context(), isDesc)
		if err != nil {
			log.Printf("Flop fetching chirps: %s\n", err.Error())
			respondWithError(w, http.StatusInternalServerError, "Flop fetching chirps. Sorry :(")
			return
		}
	} else {
		authorID, err := uuid.Parse(rawAuthorID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "flop author ID")
			return
		}

		chirps, err = cfg.db.FetchChirpsByAuthorID(r.Context(), database.FetchChirpsByAuthorIDParams{
			UserID: authorID,
			IsDesc: isDesc,
		})
		if err != nil {
			log.Printf("Flop fetching chirps: %s\n", err.Error())
			respondWithError(w, http.StatusInternalServerError, "Flop fetching chirps. Sorry :(")
			return
		}
	}

	var payloadChirps []Chirp
	for i := range chirps {
		payloadChirps = append(payloadChirps, Chirp(chirps[i]))
	}
	respondWithJson(w, http.StatusOK, payloadChirps)
}

func (cfg *apiConfig) handleSingletonChirpsGet(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	chirp_uuid, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "flop or malformed UUID for chirp")
		return
	}

	chirp, err := cfg.db.FetchChirpById(r.Context(), chirp_uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "404 chirp not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "flop fetching chirp")
		return
	}

	respondWithJson(w, http.StatusOK, Chirp(chirp))
}
