package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"

	"github.com/amstein4920/chirpy-http-server/internal/database"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits  atomic.Int32
	databaseQueries *database.Queries
	platform        string
}

func main() {
	godotenv.Load()
	platform := os.Getenv("PLATFORM")
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Failed to connect to DB")
		os.Exit(1)
	}
	dbQueries := database.New(db)
	config := apiConfig{
		databaseQueries: dbQueries,
		platform:        platform,
	}

	serveMux := http.NewServeMux()
	server := http.Server{
		Handler: serveMux,
		Addr:    ":8080",
	}

	serveMux.Handle("/app/",
		config.metricsIncrement(
			http.StripPrefix("/app",
				http.FileServer(http.Dir(".")))))

	serveMux.HandleFunc("GET /admin/metrics", config.metricsHandler)
	serveMux.HandleFunc("POST /admin/reset", config.resetHandler)

	serveMux.HandleFunc("GET /api/healthz", config.healthHandler)
	serveMux.HandleFunc("POST /api/users", config.usersHandler)
	serveMux.HandleFunc("POST /api/chirps", config.chirpsHandler)

	server.ListenAndServe()
}

func censorMessage(message string) string {
	words := strings.Split(message, " ")

	for index, word := range words {
		result := slices.IndexFunc([]string{"kerfuffle", "sharbert", "fornax"}, func(badWord string) bool { return badWord == strings.ToLower(word) })
		if result != -1 {
			words[index] = "****"
		}
	}
	return strings.Join(words, " ")
}

func respondWithError(writer http.ResponseWriter, code int, message string) {
	type returnErrorParameters struct {
		Error string `json:"error"`
	}
	errorParams := returnErrorParameters{
		Error: message,
	}

	errorResponse, err := json.Marshal(errorParams)
	if err != nil {
		fmt.Printf("Error marshaling error JSON: %s", err)
		writer.WriteHeader(500)
		return
	}
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(code)
	writer.Write(errorResponse)
}

func respondWithJSON(writer http.ResponseWriter, code int, payload interface{}) {

	validResponse, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %s", err)
		writer.WriteHeader(500)
		return
	}
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(code)
	writer.Write(validResponse)
}
