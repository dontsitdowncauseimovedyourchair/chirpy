package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/dontsitdowncauseimovedyourchair/chirpy/internal/auth"
	"github.com/dontsitdowncauseimovedyourchair/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleUsersPut(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token flop: "+err.Error())
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil || userID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "flop token")
		return
	}

	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "flop request")
		return
	}

	var req request
	err = json.Unmarshal(data, &req)
	if err != nil {
		log.Printf("flop decoding request: %s\n", err.Error())
		respondWithError(w, http.StatusBadRequest, "flop request")
		return
	}

	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("flopped hashing password: %s\n", err.Error())
		respondWithError(w, http.StatusInternalServerError, "flop")
		return
	}

	err = cfg.db.UpdateEmailPassword(r.Context(), database.UpdateEmailPasswordParams{
		ID:             userID,
		Email:          req.Email,
		HashedPassword: hashed,
	})
	if err != nil {
		log.Printf("flopped updating email and pass, %s\n", err.Error())
		respondWithError(w, http.StatusInternalServerError, "flop")
		return
	}

	user, err := cfg.db.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("flop getting user: %s", err.Error())
		respondWithError(w, http.StatusInternalServerError, "flop user")
		return
	}

	respondWithJson(w, http.StatusOK, User{
		Id:         userID,
		Created_at: user.CreatedAt,
		Updated_at: user.UpdatedAt,
		Email:      req.Email,
	})
}

func (cfg *apiConfig) handleChirpsDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token flop: "+err.Error())
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil || userID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "flop token")
		return
	}

	chirpID, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "chirp not found")
		return
	}

	chirp, err := cfg.db.FetchChirpById(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "chirp not found")
		return
	}

	if userID != chirp.UserID {
		respondWithError(w, http.StatusForbidden, "not your chirp my friendo!")
		return
	}

	err = cfg.db.DeleteChirpByID(r.Context(), chirp.ID)
	if err != nil {
		log.Printf("flop deleting chirp: %s\n", err.Error())
		respondWithError(w, http.StatusInternalServerError, "flop deleting chirp")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
