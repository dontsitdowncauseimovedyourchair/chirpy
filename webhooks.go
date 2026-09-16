package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/dontsitdowncauseimovedyourchair/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlePolkaWebhookPost(w http.ResponseWriter, r *http.Request) {
	APIKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Printf("flop getting API key: %s\n", err.Error())
		respondWithError(w, http.StatusUnauthorized, "flop API key")
		return
	}

	if APIKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "flop API key")
		return
	}

	type reqBody struct {
		Event string `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		}
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "flop data or event")
		return
	}

	var rb reqBody
	err = json.Unmarshal(data, &rb)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "flop data or event")
		return
	}

	userID, err := uuid.Parse(rb.Data.UserId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "flop user id")
		return
	}

	err = cfg.db.UpgradeUserByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "flop user id")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
