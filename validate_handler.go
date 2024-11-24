package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (config *apiConfig) validateHandler(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Body string `json:"body"`
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
	type returnParameters struct {
		CleanedBody string `json:"cleaned_body"`
	}

	censoredMessage := censorMessage(params.Body)

	validParams := returnParameters{
		CleanedBody: censoredMessage,
	}
	respondWithJSON(writer, 200, validParams)
}
