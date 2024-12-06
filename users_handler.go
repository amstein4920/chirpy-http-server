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
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type EmailPassword struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Response struct {
	User         User   `json:"user"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	IsChirpyRed  bool   `json:"is_chirpy_red"`
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
		respondWithError(writer, 500, fmt.Sprintf("User Not Created: %s", err))
		return
	}

	user := User{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed.Bool,
	}
	respondWithJSON(writer, 201, user)
}

func (config *apiConfig) usersUpdateHandler(writer http.ResponseWriter, request *http.Request) {
	accessToken, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, 401, "No Access")
		return
	}
	userId, err := auth.ValidateJWT(accessToken, config.secret)
	if err != nil {
		respondWithError(writer, 401, "No Access")
		return
	}

	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	params := parameters{}

	decoder := json.NewDecoder(request.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(writer, 500, "Invalid JSON")
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(writer, 500, "Password Failure")
	}

	dbParams := database.UpdatePassEmailParams{
		ID: userId,
		HashedPassword: sql.NullString{
			String: hashedPassword,
			Valid:  true,
		},
		Email: params.Email,
	}

	dbUser, err := config.databaseQueries.UpdatePassEmail(request.Context(), dbParams)
	if err != nil {
		respondWithError(writer, 500, "Error updating user")
	}

	user := User{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed.Bool,
	}
	respondWithJSON(writer, 200, user)
}

func decodeEmailPassword(request *http.Request) (EmailPassword, error) {
	decoder := json.NewDecoder(request.Body)
	para := EmailPassword{}
	err := decoder.Decode(&para)
	return para, err
}
