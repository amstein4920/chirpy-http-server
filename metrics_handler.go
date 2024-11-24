package main

import (
	"fmt"
	"net/http"
)

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
