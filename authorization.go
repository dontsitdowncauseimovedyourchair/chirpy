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
