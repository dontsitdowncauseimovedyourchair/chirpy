package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dontsitdowncauseimovedyourchair/chirpy/internal/auth"
	"github.com/dontsitdowncauseimovedyourchair/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleLoginPost(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	var decoded LoginRequest
	err := decoder.Decode(&decoded)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "json is flop")
		return
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

	jwt, err := auth.MakeJWT(user.ID, cfg.secret, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "flop making your token")
		return
	}

	rTkn, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "flop making refresh token")
		return
	}

	refreshToken, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     rTkn,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour).UTC(),
		RevokedAt: sql.NullTime{},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "flop creating refresh token")
		return
	}

	respondWithJson(w, http.StatusOK, User{
		Id:           user.ID,
		Created_at:   user.CreatedAt,
		Updated_at:   user.UpdatedAt,
		Email:        user.Email,
		Token:        jwt,
		RefreshToken: refreshToken.Token,
	})
}

func (cfg *apiConfig) handleRefreshPost(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token flop: "+err.Error())
		return
	}
	userID, err := cfg.db.FetchUserIDFromToken(r.Context(), token)
	if err != nil || userID == uuid.Nil {
		respondWithError(w, http.StatusUnauthorized, "you posses the token, but it is flop")
		return
	}
	jwt, err := auth.MakeJWT(userID, cfg.secret, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "what")
		return
	}
	respondWithJson(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{
		Token: jwt,
	})
}

func (cfg *apiConfig) handleRevokePost(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "token flop: "+err.Error())
		return
	}
	err = cfg.db.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		log.Printf("flop revoking: %s", err.Error())
		respondWithError(w, http.StatusInternalServerError, "revoke flop")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
