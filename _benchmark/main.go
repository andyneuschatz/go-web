package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

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

func jsonHandler(ctx *web.Ctx) web.Result {
	time.Sleep(1000 * time.Millisecond)
	ctx.Response.Header().Set(HeaderContentType, ContentTypeJSON)
	ctx.Response.Header().Set(HeaderServer, ServerName)
	json.NewEncoder(ctx.Response).Encode(&message{Message: MessageText})
	return nil
}

func jsonResultHandler(ctx *web.Ctx) web.Result {
	time.Sleep(1000 * time.Millisecond)
	return ctx.JSON().Result(&message{Message: MessageText})
}

func main() {
	app := web.New()
	app.SetPort(port())
	app.SetDiagnostics(logger.NewDiagnosticsAgentFromEnvironment())
	app.GET("/json", jsonHandler)
	app.GET("/json_result", jsonResultHandler)
	log.Fatal(app.Start())
}
