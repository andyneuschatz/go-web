package web

import (
	"net/http"

	"github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
)

// NewXMLResultProvider Creates a new JSONResults object.
func NewXMLResultProvider(diag *logger.DiagnosticsAgent, r *RequestContext) *XMLResultProvider {
	return &XMLResultProvider{diagnostics: diag, requestContext: r}
}

// XMLResultProvider are context results for api methods.
type XMLResultProvider struct {
	diagnostics    *logger.DiagnosticsAgent
	requestContext *RequestContext
}

// NotFound returns a service response.
func (xrp *XMLResultProvider) NotFound() ControllerResult {
	return &XMLResult{
		StatusCode: http.StatusNotFound,
		Response:   "Not Found",
	}
}

// NotAuthorized returns a service response.
func (xrp *XMLResultProvider) NotAuthorized() ControllerResult {
	return &XMLResult{
		StatusCode: http.StatusForbidden,
		Response:   "Not Authorized",
	}
}

// InternalError returns a service response.
func (xrp *XMLResultProvider) InternalError(err error) ControllerResult {
	if xrp.diagnostics != nil {
		if xrp.requestContext != nil {
			xrp.diagnostics.FatalWithReq(err, xrp.requestContext.Request)
		} else {
			xrp.diagnostics.FatalWithReq(err, nil)
		}
	}

	if exPtr, isException := err.(*exception.Exception); isException {
		return &XMLResult{
			StatusCode: http.StatusInternalServerError,
			Response:   exPtr,
		}
	}

	return &XMLResult{
		StatusCode: http.StatusInternalServerError,
		Response:   err.Error(),
	}
}

// BadRequest returns a service response.
func (xrp *XMLResultProvider) BadRequest(message string) ControllerResult {
	return &XMLResult{
		StatusCode: http.StatusBadRequest,
		Response:   message,
	}
}

// OK returns a service response.
func (xrp *XMLResultProvider) OK() ControllerResult {
	return &XMLResult{
		StatusCode: http.StatusOK,
		Response:   "OK!",
	}
}

// Result returns an xml response.
func (xrp *XMLResultProvider) Result(response interface{}) ControllerResult {
	return &XMLResult{
		StatusCode: http.StatusOK,
		Response:   response,
	}
}
