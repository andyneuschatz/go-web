package main

import (
	"encoding/json"
	"net/http"
	"os"
)

const (
	HEADER_CONTENT_LENGTH = "Content-Length"
	HEADER_CONTENT_TYPE   = "Content-Type"
	HEADER_SERVER         = "Server"
	HEADER_DATE           = "Date"
	SERVER_NAME           = "golang"
	APPLICATION_JSON      = "application/json; charset=UTF-8"
	MESSAGE_TEXT          = "Hello, World!"
	JSON_CONTENT_LENGTH   = "27"
)

type Message struct {
	Message string `json:"message"`
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(HEADER_CONTENT_TYPE, APPLICATION_JSON)
	w.Header().Set(HEADER_SERVER, SERVER_NAME)
	json.NewEncoder(w).Encode(&Message{Message: MESSAGE_TEXT})
}

func port() string {
	envPort := os.Getenv("PORT")
	if len(envPort) != 0 {
		return envPort
	}
	return "9090"
}

func main() {
	http.HandleFunc("/json", jsonHandler)
	http.ListenAndServe(":"+port(), nil)
}