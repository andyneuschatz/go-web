package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"strings"

	logger "github.com/blendlabs/go-logger"
)

const (
	// PostBodySize is the maximum post body size we will typically consume.
	PostBodySize = int64(1 << 26) //64mb

	// PostBodySizeMax is the absolute maximum file size the server can handle.
	PostBodySizeMax = int64(1 << 32) //enormous.

	// StringEmpty is the empty string.
	StringEmpty = ""
)

// PostedFile is a file that has been posted to an hc endpoint.
type PostedFile struct {
	Key      string
	FileName string
	Contents []byte
}

// State is the collection of state objects on a context.
type State map[string]interface{}

// NewRequestContext returns a new hc context.
func NewRequestContext(w ResponseWriter, r *http.Request, p RouteParameters) *RequestContext {
	ctx := &RequestContext{
		Response:        w,
		Request:         r,
		routeParameters: p,
		state:           State{},
	}

	return ctx
}

// RequestContext is the struct that represents the context for an hc request.
type RequestContext struct {
	//Public fields
	Response ResponseWriter
	Request  *http.Request

	postBody []byte

	//Private fields
	api                   *APIResultProvider
	view                  *ViewResultProvider
	text                  *TextResultProvider
	defaultResultProvider ControllerResultProvider
	app                   *App
	diagnostics           *logger.DiagnosticsAgent
	config                interface{}
	auth                  *SessionManager
	tx                    *sql.Tx
	state                 State
	routeParameters       RouteParameters
	statusCode            int
	contentLength         int
	requestStart          time.Time
	requestEnd            time.Time
	requestLogFormat      string
	session               *Session
}

// isolateTo isolates a request context to a transaction.
func (rc *RequestContext) isolateTo(tx *sql.Tx) *RequestContext {
	rc.tx = tx
	return rc
}

// Tx returns the transaction a request context may or may not be isolated to.
func (rc *RequestContext) Tx() *sql.Tx {
	return rc.tx
}

// Auth returns the SessionManager for the request.
func (rc *RequestContext) Auth() *SessionManager {
	return rc.auth
}

// SetAuth sets the request context auth.
func (rc *RequestContext) SetAuth(sessionManager *SessionManager) {
	rc.auth = sessionManager
}

// Session returns the session (if any) on the request.
func (rc *RequestContext) Session() *Session {
	if rc.session != nil {
		return rc.session
	}

	return nil
}

// SetSession sets the session for the request.
func (rc *RequestContext) SetSession(session *Session) {
	rc.session = session
}

// TxBegin either returns the existing (testing) transaction on the request, or calls the provider.
func (rc *RequestContext) TxBegin(provider func() (*sql.Tx, error)) (*sql.Tx, error) {
	if rc.tx != nil {
		return rc.tx, nil
	}
	return provider()
}

// TxCommit will call the committer if the request context is not isolated to a transaction already.
func (rc *RequestContext) TxCommit(commiter func() error) error {
	if rc.tx != nil {
		return nil
	}
	return commiter()
}

// TxRollback will call the rollbacker if the request context is not isolated to a transaction already.
func (rc *RequestContext) TxRollback(rollbacker func() error) error {
	if rc.tx != nil {
		return nil
	}
	return rollbacker()
}

// API returns the API result provider.
func (rc *RequestContext) API() *APIResultProvider {
	if rc.api == nil {
		rc.api = NewAPIResultProvider(rc.diagnostics, rc)
	}
	return rc.api
}

// View returns the view result provider.
func (rc *RequestContext) View() *ViewResultProvider {
	if rc.view == nil {
		rc.view = NewViewResultProvider(rc.app.diagnostics, rc.app.viewCache, rc)
	}
	return rc.view
}

// Text returns the text result provider.
func (rc *RequestContext) Text() *TextResultProvider {
	if rc.text == nil {
		rc.text = NewTextResultProvider(rc.app.diagnostics, rc)
	}
	return rc.text
}

// DefaultResultProvider returns the current result provider for the context. This is
// set by calling SetDefaultResultProvider or using one of the pre-built middleware
// steps that set it for you.
func (rc *RequestContext) DefaultResultProvider() ControllerResultProvider {
	if rc.defaultResultProvider == nil {
		rc.defaultResultProvider = NewTextResultProvider(rc.diagnostics, rc)
	}
	return rc.defaultResultProvider
}

// SetDefaultResultProvider sets the current result provider.
func (rc *RequestContext) SetDefaultResultProvider(provider ControllerResultProvider) {
	rc.defaultResultProvider = provider
}

// State returns an object in the state cache.
func (rc *RequestContext) State(key string) interface{} {
	if item, hasItem := rc.state[key]; hasItem {
		return item
	}
	return nil
}

// SetState sets the state for a key to an object.
func (rc *RequestContext) SetState(key string, value interface{}) {
	rc.state[key] = value
}

// Param returns a parameter from the request.
func (rc *RequestContext) Param(name string) string {
	if rc.routeParameters != nil {
		routeValue := rc.routeParameters.Get(name)
		if len(routeValue) > 0 {
			return routeValue
		}
	}
	if rc.Request != nil {
		if rc.Request.URL != nil {
			queryValue := rc.Request.URL.Query().Get(name)
			if len(queryValue) > 0 {
				return queryValue
			}
		}
		if rc.Request.Header != nil {
			headerValue := rc.Request.Header.Get(name)
			if len(headerValue) > 0 {
				return headerValue
			}
		}

		formValue := rc.Request.FormValue(name)
		if len(formValue) > 0 {
			return formValue
		}

		cookie, cookieErr := rc.Request.Cookie(name)
		if cookieErr == nil && len(cookie.Value) != 0 {
			return cookie.Value
		}
	}

	return ""
}

// ParamInt returns a parameter from any location as an integer.
func (rc *RequestContext) ParamInt(name string) (int, error) {
	paramValue := rc.Param(name)
	if len(paramValue) == 0 {
		return 0, parameterMissingError(name)
	}
	return strconv.Atoi(paramValue)
}

// ParamInt64 returns a parameter from any location as an int64.
func (rc *RequestContext) ParamInt64(name string) (int64, error) {
	paramValue := rc.Param(name)
	if len(paramValue) == 0 {
		return 0, parameterMissingError(name)
	}
	return strconv.ParseInt(paramValue, 10, 64)
}

// ParamFloat64 returns a parameter from any location as a float64.
func (rc *RequestContext) ParamFloat64(name string) (float64, error) {
	paramValue := rc.Param(name)
	if len(paramValue) == 0 {
		return 0, parameterMissingError(name)
	}
	return strconv.ParseFloat(paramValue, 64)
}

// ParamTime returns a parameter from any location as a time with a given format.
func (rc *RequestContext) ParamTime(name, format string) (time.Time, error) {
	paramValue := rc.Param(name)
	if len(paramValue) == 0 {
		return time.Time{}, parameterMissingError(name)
	}
	return time.Parse(format, paramValue)
}

// ParamBool returns a boolean value for a param.
func (rc *RequestContext) ParamBool(name string) (bool, error) {
	paramValue := rc.Param(name)
	if len(paramValue) == 0 {
		return false, parameterMissingError(name)
	}
	lower := strings.ToLower(paramValue)
	return lower == "true" || lower == "1" || lower == "yes", nil
}

// PostBody returns the bytes in a post body.
func (rc *RequestContext) PostBody() []byte {
	if len(rc.postBody) == 0 {
		defer rc.Request.Body.Close()
		rc.postBody, _ = ioutil.ReadAll(rc.Request.Body)
		if rc.diagnostics != nil {
			rc.diagnostics.OnEvent(logger.EventWebRequestPostBody, rc.postBody)
		}
	}

	return rc.postBody
}

// PostBodyAsString returns the post body as a string.
func (rc *RequestContext) PostBodyAsString() string {
	return string(rc.PostBody())
}

// PostBodyAsJSON reads the incoming post body (closing it) and marshals it to the target object as json.
func (rc *RequestContext) PostBodyAsJSON(response interface{}) error {
	return json.Unmarshal(rc.PostBody(), response)
}

// PostedFiles returns any files posted
func (rc *RequestContext) PostedFiles() ([]PostedFile, error) {
	var files []PostedFile

	err := rc.Request.ParseMultipartForm(PostBodySize)
	if err == nil {
		for key := range rc.Request.MultipartForm.File {
			fileReader, fileHeader, err := rc.Request.FormFile(key)
			if err != nil {
				return nil, err
			}
			bytes, err := ioutil.ReadAll(fileReader)
			if err != nil {
				return nil, err
			}
			files = append(files, PostedFile{Key: key, FileName: fileHeader.Filename, Contents: bytes})
		}
	} else {
		err = rc.Request.ParseForm()
		if err == nil {
			for key := range rc.Request.PostForm {
				if fileReader, fileHeader, err := rc.Request.FormFile(key); err == nil && fileReader != nil {
					bytes, err := ioutil.ReadAll(fileReader)
					if err != nil {
						return nil, err
					}
					files = append(files, PostedFile{Key: key, FileName: fileHeader.Filename, Contents: bytes})
				}
			}
		}
	}
	return files, nil
}

func parameterMissingError(paramName string) error {
	return fmt.Errorf("`%s` parameter is missing", paramName)
}

// RouteParamInt returns a route parameter as an integer.
func (rc *RequestContext) RouteParamInt(key string) (int, error) {
	if value, hasKey := rc.routeParameters[key]; hasKey {
		return strconv.Atoi(value)
	}
	return 0, parameterMissingError(key)
}

// RouteParamInt64 returns a route parameter as an integer.
func (rc *RequestContext) RouteParamInt64(key string) (int64, error) {
	if value, hasKey := rc.routeParameters[key]; hasKey {
		return strconv.ParseInt(value, 10, 64)
	}
	return 0, parameterMissingError(key)
}

// RouteParamFloat64 returns a route parameter as an float64.
func (rc *RequestContext) RouteParamFloat64(key string) (float64, error) {
	if value, hasKey := rc.routeParameters[key]; hasKey {
		return strconv.ParseFloat(value, 64)
	}
	return 0, parameterMissingError(key)
}

// RouteParam returns a string route parameter
func (rc *RequestContext) RouteParam(key string) (string, error) {
	if value, hasKey := rc.routeParameters[key]; hasKey {
		return value, nil
	}
	return StringEmpty, parameterMissingError(key)
}

// QueryParam returns a query parameter.
func (rc *RequestContext) QueryParam(key string) (string, error) {
	if value := rc.Request.URL.Query().Get(key); len(value) > 0 {
		return value, nil
	}
	return StringEmpty, parameterMissingError(key)
}

// QueryParamInt returns a query parameter as an integer.
func (rc *RequestContext) QueryParamInt(key string) (int, error) {
	if value := rc.Request.URL.Query().Get(key); len(value) > 0 {
		return strconv.Atoi(value)
	}
	return 0, parameterMissingError(key)
}

// QueryParamInt64 returns a query parameter as an int64.
func (rc *RequestContext) QueryParamInt64(key string) (int64, error) {
	if value := rc.Request.URL.Query().Get(key); len(value) > 0 {
		return strconv.ParseInt(value, 10, 64)
	}
	return 0, parameterMissingError(key)
}

// QueryParamFloat64 returns a query parameter as a float64.
func (rc *RequestContext) QueryParamFloat64(key string) (float64, error) {
	if value := rc.Request.URL.Query().Get(key); len(value) > 0 {
		return strconv.ParseFloat(value, 64)
	}
	return 0, parameterMissingError(key)
}

// QueryParamTime returns a query parameter as a time.Time.
func (rc *RequestContext) QueryParamTime(key, format string) (time.Time, error) {
	if value := rc.Request.URL.Query().Get(key); len(value) > 0 {
		return time.Parse(format, value)
	}
	return time.Time{}, parameterMissingError(key)
}

// HeaderParam returns a header parameter value.
func (rc *RequestContext) HeaderParam(key string) (string, error) {
	if value := rc.Request.Header.Get(key); len(value) > 0 {
		return value, nil
	}
	return StringEmpty, parameterMissingError(key)
}

// HeaderParamInt returns a header parameter value as an integer.
func (rc *RequestContext) HeaderParamInt(key string) (int, error) {
	if value := rc.Request.Header.Get(key); len(value) > 0 {
		return strconv.Atoi(value)
	}
	return 0, parameterMissingError(key)
}

// HeaderParamInt64 returns a header parameter value as an integer.
func (rc *RequestContext) HeaderParamInt64(key string) (int64, error) {
	if value := rc.Request.Header.Get(key); len(value) > 0 {
		return strconv.ParseInt(value, 10, 64)
	}
	return 0, parameterMissingError(key)
}

// HeaderParamFloat64 returns a header parameter value as an float64.
func (rc *RequestContext) HeaderParamFloat64(key string) (float64, error) {
	if value := rc.Request.Header.Get(key); len(value) > 0 {
		return strconv.ParseFloat(value, 64)
	}
	return 0, parameterMissingError(key)
}

// HeaderParamTime returns a header parameter value as an float64.
func (rc *RequestContext) HeaderParamTime(key, format string) (time.Time, error) {
	if value := rc.Request.Header.Get(key); len(value) > 0 {
		return time.Parse(format, key)
	}
	return time.Time{}, parameterMissingError(key)
}

// GetCookie returns a named cookie from the request.
func (rc *RequestContext) GetCookie(name string) *http.Cookie {
	cookie, err := rc.Request.Cookie(name)
	if err != nil {
		return nil
	}
	return cookie
}

// WriteCookie writes the cookie to the response.
func (rc *RequestContext) WriteCookie(cookie *http.Cookie) {
	http.SetCookie(rc.Response, cookie)
}

// WriteNewCookie is a helper method for WriteCookie.
func (rc *RequestContext) WriteNewCookie(name string, value string, expires *time.Time, path string, secure bool) {
	c := http.Cookie{}
	c.Name = name
	c.HttpOnly = true
	if rc.app != nil && len(rc.app.domain) > 0 {
		c.Domain = rc.app.domain
	} else {
		c.Domain = rc.Request.Host
	}
	c.Value = value
	c.Path = path
	c.Secure = secure
	if expires != nil {
		c.Expires = *expires
	}
	rc.WriteCookie(&c)
}

// ExtendCookieByDuration extends a cookie by a time duration (on the order of nanoseconds to hours).
func (rc *RequestContext) ExtendCookieByDuration(name string, duration time.Duration) {
	cookie := rc.GetCookie(name)
	cookie.Expires = cookie.Expires.Add(duration)
	rc.WriteCookie(cookie)
}

// ExtendCookie extends a cookie by years, months or days.
func (rc *RequestContext) ExtendCookie(name string, years, months, days int) {
	cookie := rc.GetCookie(name)
	cookie.Expires.AddDate(years, months, days)
	rc.WriteCookie(cookie)
}

// ExpireCookie expires a cookie.
func (rc *RequestContext) ExpireCookie(name string) {
	c := http.Cookie{}
	c.Name = name
	c.Expires = time.Now().UTC().AddDate(-1, 0, 0)
	rc.WriteCookie(&c)
}

// --------------------------------------------------------------------------------
// Diagnostics
// --------------------------------------------------------------------------------

// Diagnostics returns the diagnostics agent.
func (rc *RequestContext) Diagnostics() *logger.DiagnosticsAgent {
	return rc.diagnostics
}

// Config returns the app config.
func (rc *RequestContext) Config() interface{} {
	return rc.config
}

// --------------------------------------------------------------------------------
// Basic result providers
// --------------------------------------------------------------------------------

// Raw returns a binary response body, sniffing the content type.
func (rc *RequestContext) Raw(body []byte) *RawResult {
	sniffedContentType := http.DetectContentType(body)
	return rc.RawWithContentType(sniffedContentType, body)
}

// RawWithContentType returns a binary response with a given content type.
func (rc *RequestContext) RawWithContentType(contentType string, body []byte) *RawResult {
	return &RawResult{ContentType: contentType, Body: body}
}

// JSON returns a basic json result.
func (rc *RequestContext) JSON(object interface{}) *JSONResult {
	return &JSONResult{
		StatusCode: http.StatusOK,
		Response:   object,
	}
}

// NoContent returns a service response.
func (rc *RequestContext) NoContent() *NoContentResult {
	return &NoContentResult{}
}

// Static returns a static result.
func (rc *RequestContext) Static(filePath string) *StaticResult {
	return NewStaticResultForSingleFile(filePath)
}

// Redirect returns a redirect result.
func (rc *RequestContext) Redirect(path string) *RedirectResult {
	return &RedirectResult{
		RedirectURI: path,
	}
}

// Redirectf returns a redirect result.
func (rc *RequestContext) Redirectf(format string, args ...interface{}) *RedirectResult {
	return &RedirectResult{
		RedirectURI: fmt.Sprintf(format, args...),
	}
}

// RedirectWithMethodf returns a redirect result with a given method.
func (rc *RequestContext) RedirectWithMethodf(method, format string, args ...interface{}) *RedirectResult {
	return &RedirectResult{
		Method:      method,
		RedirectURI: fmt.Sprintf(format, args...),
	}
}

// --------------------------------------------------------------------------------
// Stats Methods used for logging.
// --------------------------------------------------------------------------------

// StatusCode returns the status code for the request, this is used for logging.
func (rc *RequestContext) getLoggedStatusCode() int {
	return rc.statusCode
}

// SetStatusCode sets the status code for the request, this is used for logging.
func (rc *RequestContext) setLoggedStatusCode(code int) {
	rc.statusCode = code
}

// ContentLength returns the content length for the request, this is used for logging.
func (rc *RequestContext) getLoggedContentLength() int {
	return rc.contentLength
}

// SetContentLength sets the content length, this is used for logging.
func (rc *RequestContext) setLoggedContentLength(length int) {
	rc.contentLength = length
}

// OnRequestStart will mark the start of request timing.
func (rc *RequestContext) onRequestStart() {
	rc.requestStart = time.Now().UTC()
}

// Start returns the request start time.
func (rc RequestContext) Start() time.Time {
	return rc.requestStart
}

// OnRequestEnd will mark the end of request timing.
func (rc *RequestContext) onRequestEnd() {
	rc.requestEnd = time.Now().UTC()
}

// Elapsed is the time delta between start and end.
func (rc *RequestContext) Elapsed() time.Duration {
	if !rc.requestEnd.IsZero() {
		return rc.requestEnd.Sub(rc.requestStart)
	}
	return time.Now().UTC().Sub(rc.requestStart)
}