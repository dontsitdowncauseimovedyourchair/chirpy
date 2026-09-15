package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dontsitdowncauseimovedyourchair/chirpy/internal/auth"
)

func (cfg *apiConfig) handleLoginPost(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	var decoded LoginRequest
	err := decoder.Decode(&decoded)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "json is flop")
		return
	}
	if decoded.ExpiresInSeconds == 0 {
		decoded.ExpiresInSeconds = 3600
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), decoded.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password")
		return
	}

	check, err := auth.CheckPassword(decoded.Password, user.HashedPassword)
	if err != nil || check == false {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password")
		return
	}

	jwt, err := auth.MakeJWT(user.ID, cfg.secret, time.Duration(decoded.ExpiresInSeconds)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "flop making your token")
		return
	}

	respondWithJson(w, http.StatusOK, User{
		Id:         user.ID,
		Created_at: user.CreatedAt,
		Updated_at: user.UpdatedAt,
		Email:      user.Email,
		Token:      jwt,
	})
}
