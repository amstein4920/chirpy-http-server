package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (config *apiConfig) usersHandler(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(request.Body)
	para := parameters{}
	err := decoder.Decode(&para)
	if err != nil {
		fmt.Printf("Invalid JSON: %s", err)
		writer.WriteHeader(500)
		return
	}

	dbUser, err := config.databaseQueries.CreateUser(request.Context(), para.Email)
	if err != nil {
		fmt.Printf("User Not Created: %s", err)
		writer.WriteHeader(500)
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	respondWithJSON(writer, 201, user)
}
