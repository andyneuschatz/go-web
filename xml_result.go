package web

// XMLResult is a json result.
type XMLResult struct {
	StatusCode int
	Response   interface{}
}

// Render renders the result
func (ar *XMLResult) Render(rc *RequestContext) error {
	return WriteXML(rc.Response, rc.Request, ar.StatusCode, ar.Response)
}
