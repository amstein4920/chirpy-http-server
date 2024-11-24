package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/amstein4920/chirpy-http-server/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (config *apiConfig) chirpsHandler(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	decoder := json.NewDecoder(request.Body)
	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		fmt.Printf("Invalid JSON: %s", err)
		writer.WriteHeader(500)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(writer, 400, "Chirp is too long")
		return
	}

	params.Body = censorMessage(params.Body)

	dbChirp, err := config.databaseQueries.CreateChirp(request.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: params.UserID,
	})
	if err != nil {
		fmt.Printf("Chirp not created: %s", err)
		writer.WriteHeader(500)
		return
	}

	returnChirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	respondWithJSON(writer, 201, returnChirp)
}
