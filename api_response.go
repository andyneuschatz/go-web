package web

import "github.com/blendlabs/go-exception"

// APIResponseMeta is the meta component of a service response.
type APIResponseMeta struct {
	StatusCode int
	Message    string               `json:",omitempty"`
	Exception  *exception.Exception `json:",omitempty"`
}

// APIResponse is the standard API response format.
type APIResponse struct {
	Meta     *APIResponseMeta
	Response interface{}
}
