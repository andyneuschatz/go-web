package web

import (
	"fmt"
	"net/http"

	logger "github.com/blendlabs/go-logger"
)

// NewTextResultProvider returns a new text result provider.
func NewTextResultProvider(diag *logger.DiagnosticsAgent, r *RequestContext) *TextResultProvider {
	return &TextResultProvider{diagnostics: diag, requestContext: r}
}

// TextResultProvider is the default response provider if none is specified.
type TextResultProvider struct {
	diagnostics    *logger.DiagnosticsAgent
	requestContext *RequestContext
}

// NotFound returns a text response.
func (trp *TextResultProvider) NotFound() ControllerResult {
	return &RawResult{
		StatusCode:  http.StatusNotFound,
		ContentType: ContentTypeText,
		Body:        []byte("Not Found"),
	}
}

// NotAuthorized returns a text response.
func (trp *TextResultProvider) NotAuthorized() ControllerResult {
	return &RawResult{
		StatusCode:  http.StatusForbidden,
		ContentType: ContentTypeText,
		Body:        []byte("Not Authorized"),
	}
}

// InternalError returns a text response.
func (trp *TextResultProvider) InternalError(err error) ControllerResult {
	if trp.diagnostics != nil {
		if trp.requestContext != nil {
			trp.diagnostics.FatalWithReq(err, trp.requestContext.Request)
		} else {
			trp.diagnostics.FatalWithReq(err, nil)
		}
	}

	if err != nil {
		return &RawResult{
			StatusCode:  http.StatusForbidden,
			ContentType: ContentTypeText,
			Body:        []byte(err.Error()),
		}
	}

	return &RawResult{
		StatusCode:  http.StatusForbidden,
		ContentType: ContentTypeText,
		Body:        []byte("An internal server error occurred."),
	}
}

// BadRequest returns a text response.
func (trp *TextResultProvider) BadRequest(message string) ControllerResult {
	if len(message) > 0 {
		return &RawResult{
			StatusCode:  http.StatusBadRequest,
			ContentType: ContentTypeText,
			Body:        []byte(fmt.Sprintf("Bad Request: %s", message)),
		}
	}
	return &RawResult{
		StatusCode:  http.StatusBadRequest,
		ContentType: ContentTypeText,
		Body:        []byte("Bad Request"),
	}
}

// Text returns a plaintext result.
func (trp *TextResultProvider) Text(message string) ControllerResult {
	return &RawResult{
		StatusCode:  http.StatusOK,
		ContentType: ContentTypeText,
		Body:        []byte(message),
	}
}
