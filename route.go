package web

import "net/http"

// Handler is the most basic route handler.
type Handler func(http.ResponseWriter, *http.Request, *Route, RouteParameters)

// PanicHandler is a handler for panics that also takes an error.
type PanicHandler func(http.ResponseWriter, *http.Request, interface{})

// Route is an entry in the route tree.
type Route struct {
	Handler
	Method string
	Path   string
	Params []string
}
