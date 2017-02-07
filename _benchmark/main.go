package main

import (
	"encoding/json"
	"log"
	"os"

	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-web"
)

const (
	// ContentTypeJSON is the json content type.
	ContentTypeJSON = "application/json; charset=UTF-8"
	// HeaderContentLength is a header.
	HeaderContentLength = "Content-Length"
	// HeaderContentType is a header.
	HeaderContentType = "Content-Type"
	// HeaderServer is a header.
	HeaderServer = "Server"
	// ServerName is a header.
	ServerName = "golang"
	// MessageText is a string.
	MessageText = "Hello, World!"
)

var (
	// MessageBytes is the raw serialized message.
	MessageBytes = []byte(`{"message":"Hello, World!"}`)
)

type message struct {
	Message string `json:"message"`
}

func port() string {
	envPort := os.Getenv("PORT")
	if len(envPort) != 0 {
		return envPort
	}
	return "8080"
}

func jsonHandler(r *web.RequestContext) web.ControllerResult {
	r.Response.Header().Set(HeaderContentType, ContentTypeJSON)
	r.Response.Header().Set(HeaderServer, ServerName)
	json.NewEncoder(r.Response).Encode(&message{Message: MessageText})
	return nil
}

func main() {
	app := web.New()
	app.SetPort(port())
	app.SetDiagnostics(logger.NewDiagnosticsAgentFromEnvironment())
	app.GET("/json", jsonHandler)
	log.Fatal(app.Start())
}
