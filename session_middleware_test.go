package web

import (
	"net/http"
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestSessionAware(t *testing.T) {
	assert := assert.New(t)

	sessionID := String.SecureRandom(64)

	var sessionWasSet bool
	app := New()

	app.GET("/", func(r *RequestContext) ControllerResult {
		sessionWasSet = r.Session() != nil
		return r.Text().Text("COOL")
	}, SessionAware)

	app.Auth().SessionCache().Add(&Session{
		SessionID: sessionID,
	})

	meta, err := app.Mock().WithPathf("/").WithHeader(app.Auth().SessionParamName(), sessionID).ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)
	assert.True(sessionWasSet)

	unsetMeta, err := app.Mock().WithPathf("/").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, unsetMeta.StatusCode)
	assert.False(sessionWasSet)
}

func TestSessionRequired(t *testing.T) {
	assert := assert.New(t)

	sessionID := String.SecureRandom(64)

	var sessionWasSet bool
	app := New()

	app.GET("/", func(r *RequestContext) ControllerResult {
		sessionWasSet = r.Session() != nil
		return r.Text().Text("COOL")
	}, SessionRequired)

	app.Auth().SessionCache().Add(&Session{
		SessionID: sessionID,
	})

	unsetMeta, err := app.Mock().WithPathf("/").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusForbidden, unsetMeta.StatusCode)
	assert.False(sessionWasSet)

	meta, err := app.Mock().WithPathf("/").WithHeader(app.Auth().SessionParamName(), sessionID).ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)
	assert.True(sessionWasSet)
}

func TestSessionRequiredCustomParamName(t *testing.T) {
	assert := assert.New(t)

	sessionID := String.SecureRandom(64)

	var sessionWasSet bool
	app := New()
	app.Auth().SetSessionParamName("web_auth")

	app.GET("/", func(r *RequestContext) ControllerResult {
		sessionWasSet = r.Session() != nil
		return r.Text().Text("COOL")
	}, SessionRequired)

	app.Auth().SessionCache().Add(&Session{
		SessionID: sessionID,
	})

	unsetMeta, err := app.Mock().WithPathf("/").ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusForbidden, unsetMeta.StatusCode)
	assert.False(sessionWasSet)

	meta, err := app.Mock().WithPathf("/").WithHeader(app.Auth().SessionParamName(), sessionID).ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusOK, meta.StatusCode)
	assert.True(sessionWasSet)

	meta, err = app.Mock().WithPathf("/").WithHeader(DefaultSessionParamName, sessionID).ExecuteWithMeta()
	assert.Nil(err)
	assert.Equal(http.StatusForbidden, meta.StatusCode)
	assert.True(sessionWasSet)
}
