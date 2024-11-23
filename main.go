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
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Failed to connect to DB")
		os.Exit(1)
	}
	dbQueries := database.New(db)
	config := apiConfig{
		databaseQueries: dbQueries,
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

	serveMux.HandleFunc("GET /api/healthz", config.healthHandler)
	serveMux.HandleFunc("GET /admin/metrics", config.metricsHandler)

	serveMux.HandleFunc("POST /admin/reset", config.resetHandler)
	serveMux.HandleFunc("POST /api/validate_chirp", config.validateHandler)

	server.ListenAndServe()
}

func (config *apiConfig) healthHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}

func (config *apiConfig) metricsIncrement(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		config.fileserverHits.Add(1)
		next.ServeHTTP(writer, request)
	})
}

func (config *apiConfig) metricsHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Add("Content-Type", "text/html")
	writer.WriteHeader(200)
	writer.Write([]byte(fmt.Sprintf(`
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>`,
		config.fileserverHits.Load())))
}

func (config *apiConfig) resetHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(200)
	config.fileserverHits.Store(0)
}

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
