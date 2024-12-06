package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/amstein4920/chirpy-http-server/internal/auth"
	"github.com/google/uuid"
)

func (config *apiConfig) webhooksHandler(writer http.ResponseWriter, request *http.Request) {
	apiKeyFromHeader, err := auth.GetAPIKey(request.Header)
	if err != nil {
		respondWithError(writer, 401, "Auth Not Found")
		return
	}
	if apiKeyFromHeader != config.polkaKey {
		respondWithError(writer, 401, "Auth Not Found")
		return
	}

	type inputParams struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	params := inputParams{}
	decoder := json.NewDecoder(request.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Invalid JSON: %s", err))
		return
	}
	if params.Event != "user.upgraded" {
		writer.WriteHeader(http.StatusNoContent)
		return
	}
	err = config.databaseQueries.UpdateUsersRed(request.Context(), params.Data.UserID)
	if err != nil {
		respondWithError(writer, http.StatusNotFound, "User not found")
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}
