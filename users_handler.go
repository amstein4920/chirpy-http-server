package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/amstein4920/chirpy-http-server/internal/auth"
	"github.com/amstein4920/chirpy-http-server/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type EmailPassword struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Response struct {
	User         User   `json:"user"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (config *apiConfig) usersHandler(writer http.ResponseWriter, request *http.Request) {
	para, err := decodeEmailPassword(request)
	if err != nil {
		fmt.Printf("Invalid JSON: %s", err)
		writer.WriteHeader(500)
		return
	}

	hashedPass, err := auth.HashPassword(para.Password)
	if err != nil {
		fmt.Printf("Password Failure: %s", err)
		writer.WriteHeader(500)
		return
	}

	dbParams := database.CreateUserParams{
		Email: para.Email,
		HashedPassword: sql.NullString{
			String: hashedPass,
			Valid:  true,
		},
	}

	dbUser, err := config.databaseQueries.CreateUser(request.Context(), dbParams)
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

func decodeEmailPassword(request *http.Request) (EmailPassword, error) {
	decoder := json.NewDecoder(request.Body)
	para := EmailPassword{}
	err := decoder.Decode(&para)
	return para, err
}
