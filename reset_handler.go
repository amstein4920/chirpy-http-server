package main

import (
	"fmt"
	"net/http"
)

func (config *apiConfig) resetHandler(writer http.ResponseWriter, request *http.Request) {
	if config.platform != "dev" {
		fmt.Println("Not Allowed")
		writer.WriteHeader(403)
		return
	}
	config.databaseQueries.DelUsers(request.Context())
	writer.WriteHeader(200)
	config.fileserverHits.Store(0)
}
