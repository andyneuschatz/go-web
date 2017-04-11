package web

import (
	"database/sql"
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestAuthManagerLogin(t *testing.T) {
	assert := assert.New(t)

	app := New()
	rc, _ := app.Mock().Ctx(nil)

	am := NewAuthManager()
	session, err := am.Login(1, rc)
	assert.Nil(err)

	rc2, err := app.Mock().WithHeader(am.SessionParamName(), session.SessionID).Ctx(nil)
	assert.Nil(err)

	valid, err := am.VerifySession(rc2)
	assert.Nil(err)
	assert.NotNil(valid)
	assert.Equal(1, valid.UserID)
}

func TestAuthManagerLoginSecure(t *testing.T) {
	assert := assert.New(t)

	app := New()
	rc, _ := app.Mock().Ctx(nil)

	am := NewAuthManager()
	am.SetSecret(GenerateSHA512Key())
	session, err := am.Login(1, rc)
	assert.Nil(err)

	secureSessionID, err := EncodeSignSessionID(session.SessionID, am.Secret())
	assert.Nil(err)

	rc2, err := app.Mock().WithHeader(am.SessionParamName(), session.SessionID).WithHeader(am.SecureSessionParamName(), secureSessionID).Ctx(nil)
	assert.Nil(err)

	valid, err := am.VerifySession(rc2)
	assert.Nil(err)
	assert.NotNil(valid)
	assert.Equal(1, valid.UserID)
}

func TestAuthManagerLoginWithPersist(t *testing.T) {
	assert := assert.New(t)

	sessions := map[string]*Session{}

	app := New()
	rc, _ := app.Mock().Ctx(nil)

	didCallPersist := false
	am := NewAuthManager()
	am.SetPersistHandler(func(c *Ctx, s *Session, tx *sql.Tx) error {
		didCallPersist = true
		sessions[s.SessionID] = s
		return nil
	})

	session, err := am.Login(1, rc)
	assert.Nil(err)
	assert.True(didCallPersist)

	am2 := NewAuthManager()
	am2.SetFetchHandler(func(sid string, tx *sql.Tx) (*Session, error) {
		return sessions[sid], nil
	})

	rc2, err := app.Mock().WithHeader(am.SessionParamName(), session.SessionID).Ctx(nil)
	assert.Nil(err)

	valid, err := am2.VerifySession(rc2)
	assert.Nil(err)
	assert.NotNil(valid)
	assert.Equal(1, valid.UserID)
}

func TestAuthManagerVerifySessionWithFetch(t *testing.T) {
	assert := assert.New(t)

	app := New()

	sessions := map[string]*Session{}

	didCallHandler := false

	am := NewAuthManager()
	am.SetFetchHandler(func(sessionID string, tx *sql.Tx) (*Session, error) {
		didCallHandler = true
		return sessions[sessionID], nil
	})
	sessionID := NewSessionID()
	sessions[sessionID] = NewSession(1, sessionID)

	rc2, err := app.Mock().WithHeader(am.SessionParamName(), sessionID).Ctx(nil)
	assert.Nil(err)

	valid, err := am.VerifySession(rc2)
	assert.Nil(err)
	assert.Equal(sessionID, valid.SessionID)
	assert.Equal(1, valid.UserID)
	assert.True(didCallHandler)

	rc3, err := app.Mock().WithHeader(am.SessionParamName(), NewSessionID()).Ctx(nil)
	assert.Nil(err)

	invalid, err := am.VerifySession(rc3)
	assert.Nil(err)
	assert.Nil(invalid)
}

func TestAuthManagerIsCookieSecure(t *testing.T) {
	assert := assert.New(t)
	sm := NewAuthManager()
	assert.False(sm.IsCookieHTTPSOnly())
	sm.SetCookieAsHTTPSOnly(true)
	assert.True(sm.IsCookieHTTPSOnly())
	sm.SetCookieAsHTTPSOnly(false)
	assert.False(sm.IsCookieHTTPSOnly())
}
