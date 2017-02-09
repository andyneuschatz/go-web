package web

import (
	"net/http"

	"github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
)

// NewJSONResultProvider Creates a new JSONResults object.
func NewJSONResultProvider(diag *logger.DiagnosticsAgent, r *RequestContext) *JSONResultProvider {
	return &JSONResultProvider{diagnostics: diag, requestContext: r}
}

// JSONResultProvider are context results for api methods.
type JSONResultProvider struct {
	diagnostics    *logger.DiagnosticsAgent
	requestContext *RequestContext
}

// NotFound returns a service response.
func (jrp *JSONResultProvider) NotFound() ControllerResult {
	return &JSONResult{
		StatusCode: http.StatusNotFound,
		Response:   "Not Found",
	}
}

// NotAuthorized returns a service response.
func (jrp *JSONResultProvider) NotAuthorized() ControllerResult {
	return &JSONResult{
		StatusCode: http.StatusForbidden,
		Response:   "Not Authorized",
	}
}

// InternalError returns a service response.
func (jrp *JSONResultProvider) InternalError(err error) ControllerResult {
	if jrp.diagnostics != nil {
		if jrp.requestContext != nil {
			jrp.diagnostics.FatalWithReq(err, jrp.requestContext.Request)
		} else {
			jrp.diagnostics.FatalWithReq(err, nil)
		}
	}

	if exPtr, isException := err.(*exception.Exception); isException {
		return &JSONResult{
			StatusCode: http.StatusInternalServerError,
			Response:   exPtr,
		}
	}

	return &JSONResult{
		StatusCode: http.StatusInternalServerError,
		Response:   err.Error(),
	}
}

// BadRequest returns a service response.
func (jrp *JSONResultProvider) BadRequest(message string) ControllerResult {
	return &JSONResult{
		StatusCode: http.StatusBadRequest,
		Response:   message,
	}
}

// OK returns a service response.
func (jrp *JSONResultProvider) OK() ControllerResult {
	return &JSONResult{
		StatusCode: http.StatusOK,
		Response:   "OK!",
	}
}

// Result returns a json response.
func (jrp *JSONResultProvider) Result(response interface{}) ControllerResult {
	return &JSONResult{
		StatusCode: http.StatusOK,
		Response:   response,
	}
}
