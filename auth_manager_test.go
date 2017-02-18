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

	sm := NewAuthManager()
	session, err := sm.Login(1, rc)
	assert.Nil(err)

	valid, err := sm.VerifySession(session.SessionID, nil)
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
	sm := NewAuthManager()
	sm.SetPersistHandler(func(c *Ctx, s *Session, tx *sql.Tx) error {
		didCallPersist = true
		sessions[s.SessionID] = s
		return nil
	})

	session, err := sm.Login(1, rc)
	assert.Nil(err)
	assert.True(didCallPersist)

	sm2 := NewAuthManager()
	sm2.SetFetchHandler(func(sid string, tx *sql.Tx) (*Session, error) {
		return sessions[sid], nil
	})

	valid, err := sm2.VerifySession(session.SessionID, nil)
	assert.Nil(err)
	assert.NotNil(valid)
	assert.Equal(1, valid.UserID)
}

func TestAuthManagerVerifySession(t *testing.T) {
	assert := assert.New(t)

	sm := NewAuthManager()
	sessionID := NewSessionID()
	sm.sessionCache.Add(NewSession(1, sessionID))

	valid, err := sm.VerifySession(sessionID, nil)
	assert.Nil(err)
	assert.Equal(sessionID, valid.SessionID)
	assert.Equal(1, valid.UserID)

	invalid, err := sm.VerifySession(NewSessionID(), nil)
	assert.Nil(err) // we do not return an error on miss if no fetch handler is configured.
	assert.Nil(invalid)
}

func TestAuthManagerVerifySessionWithFetch(t *testing.T) {
	assert := assert.New(t)

	sessions := map[string]*Session{}

	didCallHandler := false

	sm := NewAuthManager()
	sm.SetFetchHandler(func(sessionID string, tx *sql.Tx) (*Session, error) {
		didCallHandler = true
		return sessions[sessionID], nil
	})
	sessionID := NewSessionID()
	sessions[sessionID] = NewSession(1, sessionID)

	valid, err := sm.VerifySession(sessionID, nil)
	assert.Nil(err)
	assert.Equal(sessionID, valid.SessionID)
	assert.Equal(1, valid.UserID)
	assert.True(didCallHandler)

	invalid, err := sm.VerifySession(NewSessionID(), nil)
	assert.Nil(err)
	assert.Nil(invalid)
}

func TestAuthManagerIsCookieSecure(t *testing.T) {
	assert := assert.New(t)
	sm := NewAuthManager()
	assert.False(sm.IsCookieSecure())
	sm.SetCookieAsSecure(true)
	assert.True(sm.IsCookieSecure())
	sm.SetCookieAsSecure(false)
	assert.False(sm.IsCookieSecure())
}
