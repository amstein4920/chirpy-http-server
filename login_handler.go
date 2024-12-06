package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/amstein4920/chirpy-http-server/internal/auth"
	"github.com/amstein4920/chirpy-http-server/internal/database"
)

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
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(writer, 401, "Couldn't acquire refresh")
		return
	}

	user := User{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed.Bool,
	}

	respondWithJSON(writer, 200, Response{
		User:         user,
		Token:        accessToken,
		RefreshToken: refreshToken,
		IsChirpyRed:  dbUser.IsChirpyRed.Bool,
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
		return
	}
	err = config.databaseQueries.UpdateRevocation(request.Context(), token)
	if err != nil {
		respondWithError(writer, 410, "Failure")
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
