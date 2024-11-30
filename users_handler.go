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

func (config *apiConfig) loginHandler(writer http.ResponseWriter, request *http.Request) {
	para, err := decodeEmailPassword(request)
	if err != nil {
		fmt.Printf("Invalid JSON: %s", err)
		writer.WriteHeader(500)
		return
	}

	dbUser, err := config.databaseQueries.UserPassword(request.Context(), para.Email)
	if err != nil {
		fmt.Println("Incorrect email or password")
		writer.WriteHeader(401)
		return
	}

	err = auth.CheckPasswordHash(para.Password, dbUser.HashedPassword.String)
	if err != nil {
		fmt.Println("Incorrect email or password")
		writer.WriteHeader(401)
		return
	}

	accessToken, err := auth.MakeJWT(dbUser.ID, config.secret, time.Hour)
	if err != nil {
		respondWithError(writer, 401, "Couldn't access JWT")
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(writer, 401, "Couldn't acquire refresh")
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	respondWithJSON(writer, 200, Response{
		User:         user,
		Token:        accessToken,
		RefreshToken: refreshToken,
	})

	config.databaseQueries.CreateRefresh(request.Context(), database.CreateRefreshParams{
		Token:  refreshToken,
		UserID: user.ID,
	})
}

func (config *apiConfig) refreshHandler(writer http.ResponseWriter, request *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, "Couldn't find token")
		return
	}

	userId, err := config.databaseQueries.CheckRefresh(request.Context(), refreshToken)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Couldn't get user for refresh token")
		return
	}

	accessToken, err := auth.MakeJWT(
		userId,
		config.secret,
		time.Hour,
	)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Couldn't validate token")
		return
	}

	respondWithJSON(writer, http.StatusOK, response{
		Token: accessToken,
	})
}

func (config *apiConfig) revokeHandler(writer http.ResponseWriter, request *http.Request) {
	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, 401, err.Error())
	}
	err = config.databaseQueries.UpdateRevocation(request.Context(), token)
	if err != nil {
		respondWithError(writer, 410, "Failure")
	}

	writer.WriteHeader(http.StatusNoContent)
}

func decodeEmailPassword(request *http.Request) (EmailPassword, error) {
	decoder := json.NewDecoder(request.Body)
	para := EmailPassword{}
	err := decoder.Decode(&para)
	return para, err
}
