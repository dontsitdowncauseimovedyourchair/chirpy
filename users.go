package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/dontsitdowncauseimovedyourchair/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handleUsersPost(w http.ResponseWriter, r *http.Request) {
	type requestForm struct {
		Email string `json:"email"`
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Flop request")
		return
	}

	var form requestForm
	err = json.Unmarshal(data, &form)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Flop request, should only contain email field")
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     form.Email,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Flop on server registering email. Sorry :(")
		log.Println(err.Error())
		return
	}

	respondWithJson(w, http.StatusCreated, User{
		user.ID,
		user.CreatedAt,
		user.UpdatedAt,
		user.Email,
	})
}
