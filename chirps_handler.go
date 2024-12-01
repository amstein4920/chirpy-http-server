package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/amstein4920/chirpy-http-server/internal/auth"
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
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, 401, err.Error())
		return
	}

	userId, err := auth.ValidateJWT(token, config.secret)
	if err != nil {
		respondWithError(writer, 401, "Unauthorized")
		return
	}

	decoder := json.NewDecoder(request.Body)
	params := parameters{}

	err = decoder.Decode(&params)
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
		UserID: userId,
	})
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Chirp not created: %s", err.Error()))
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

func (config *apiConfig) allChirpsHandler(writer http.ResponseWriter, request *http.Request) {
	dbChirps, err := config.databaseQueries.AllChirps(request.Context())
	if err != nil {
		fmt.Printf("Chirps not retrieved: %s", err)
		writer.WriteHeader(500)
		return
	}

	returnChirps := []Chirp{}

	for _, dbChirp := range dbChirps {
		returnChirps = append(returnChirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		})
	}

	respondWithJSON(writer, 200, returnChirps)
}

func (config *apiConfig) singleChirpsHandler(writer http.ResponseWriter, request *http.Request) {
	id, err := uuid.Parse(request.PathValue("id"))
	if err != nil {
		fmt.Printf("Invalid ID: %s", err)
		writer.WriteHeader(500)
		return
	}
	dbChirp, err := config.databaseQueries.SingleChirp(request.Context(), id)
	if err != nil {
		writer.WriteHeader(404)
		return
	}

	returnChirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	respondWithJSON(writer, 200, returnChirp)
}
