package main

import (
	"log"
	"os"

	"net/http"

	"github.com/labstack/echo"
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

func jsonHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, &message{Message: MessageText})
}

func main() {
	app := echo.New()

	app.GET("/json", jsonHandler)
	log.Fatal(app.Start(":" + port()))
}
