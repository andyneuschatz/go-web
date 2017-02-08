package web

import (
	"database/sql"
	"net/url"
	"time"
)

const (
	// DefaultSessionParamName is the default name of the field (header, cookie, or querystring) that needs to have the sessionID on it.
	DefaultSessionParamName = "__auth_token__"

	// SessionLockFree is a lock-free policy.
	SessionLockFree = 0

	// SessionReadLock is a lock policy that acquires a read lock on session.
	SessionReadLock = 1

	// SessionReadWriteLock is a lock policy that acquires both a read and a write lock on session.
	SessionReadWriteLock = 2
)

// NewSessionID returns a new session id.
// It is not a uuid; session ids are generated using a secure random source.
// SessionIDs are generally 64 bytes.
func NewSessionID() string {
	return String.SecureRandom(64)
}

// NewSessionManager returns a new session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessionCache:                NewSessionCache(),
		sessionCookieIsSessionBound: true,
		sessionParamName:            DefaultSessionParamName,
	}
}

// SessionManager is a manager for sessions.
type SessionManager struct {
	sessionCache         *SessionCache
	persistHandler       func(*RequestContext, *Session, *sql.Tx) error
	fetchHandler         func(sessionID string, tx *sql.Tx) (*Session, error)
	removeHandler        func(sessionID string, tx *sql.Tx) error
	validateHandler      func(*Session, *sql.Tx) error
	loginRedirectHandler func(*url.URL) *url.URL
	sessionParamName     string

	sessionCookieIsSessionBound  bool
	sessionCookieIsSecure        *bool
	sessionCookieTimeoutProvider func(rc *RequestContext) *time.Time
}

// SetCookiesAsSessionBound sets the session issued cookies to be deleted after the browser closes.
func (sm *SessionManager) SetCookiesAsSessionBound() {
	sm.sessionCookieIsSessionBound = true
	sm.sessionCookieTimeoutProvider = nil
}

// SetCookieTimeout sets the cookies to the given timeout.
func (sm *SessionManager) SetCookieTimeout(timeoutProvider func(rc *RequestContext) *time.Time) {
	sm.sessionCookieIsSessionBound = false
	sm.sessionCookieTimeoutProvider = timeoutProvider
}

// SetCookieAsSecure overrides defaults when determining if we should use the HTTPS only cooikie option.
// The default depends on the app configuration (if tls is configured and enabled).
func (sm *SessionManager) SetCookieAsSecure(isSecure bool) {
	sm.sessionCookieIsSecure = &isSecure
}

// SessionParamName returns the session param name.
func (sm *SessionManager) SessionParamName() string {
	return sm.sessionParamName
}

// SetSessionParamName sets the session param name.
func (sm *SessionManager) SetSessionParamName(paramName string) {
	sm.sessionParamName = paramName
}

// SetPersistHandler sets the persist handler
func (sm *SessionManager) SetPersistHandler(handler func(*RequestContext, *Session, *sql.Tx) error) {
	sm.persistHandler = handler
}

// SetFetchHandler sets the fetch handler
func (sm *SessionManager) SetFetchHandler(handler func(sessionID string, tx *sql.Tx) (*Session, error)) {
	sm.fetchHandler = handler
}

// SetRemoveHandler sets the remove handler.
func (sm *SessionManager) SetRemoveHandler(handler func(sessionID string, tx *sql.Tx) error) {
	sm.removeHandler = handler
}

// SetValidateHandler sets the validate handler.
func (sm *SessionManager) SetValidateHandler(handler func(*Session, *sql.Tx) error) {
	sm.validateHandler = handler
}

// SetLoginRedirectHandler sets the handler to determin where to redirect on not authorized attempts.
// It should return (nil) if you want to just show the `not_authorized` template.
func (sm *SessionManager) SetLoginRedirectHandler(handler func(*url.URL) *url.URL) {
	sm.loginRedirectHandler = handler
}

// SessionCache returns the session cache.
func (sm SessionManager) SessionCache() *SessionCache {
	return sm.sessionCache
}

// Login logs a userID in.
func (sm *SessionManager) Login(userID int64, context *RequestContext) (*Session, error) {
	sessionID := NewSessionID()
	session := NewSession(userID, sessionID)

	var err error
	if sm.persistHandler != nil {
		err = sm.persistHandler(context, session, context.Tx())
		if err != nil {
			return nil, err
		}
	}

	sm.sessionCache.Add(session)
	sm.InjectSessionCookie(context, sessionID)
	return session, nil
}

// InjectSessionCookie injects a session cookie into the context.
func (sm *SessionManager) InjectSessionCookie(context *RequestContext, sessionID string) {
	if context != nil {
		if sm.sessionCookieIsSessionBound {
			context.WriteNewCookie(sm.sessionParamName, sessionID, nil, "/", sm.IsCookieSecure())
		} else if sm.sessionCookieTimeoutProvider != nil {
			context.WriteNewCookie(sm.sessionParamName, sessionID, sm.sessionCookieTimeoutProvider(context), "/", sm.IsCookieSecure())
		}
	}
}

// IsCookieSecure returns if the session cookie is configured to be secure only.
func (sm *SessionManager) IsCookieSecure() bool {
	return sm.sessionCookieIsSecure != nil && *sm.sessionCookieIsSecure
}

// Logout un-authenticates a session.
func (sm *SessionManager) Logout(userID int64, sessionID string, context *RequestContext) error {
	sm.sessionCache.Expire(sessionID)

	if context != nil {
		context.ExpireCookie(sm.sessionParamName)
	}
	if sm.removeHandler != nil {
		if context != nil {
			return sm.removeHandler(sessionID, context.Tx())
		}
		return sm.removeHandler(sessionID, nil)
	}
	return nil
}

// ReadSessionID reads a session id from a given request context.
func (sm *SessionManager) ReadSessionID(context *RequestContext) string {
	return context.GetCookie(sm.sessionParamName).Value
}

// VerifySession checks a sessionID to see if it's valid.
func (sm *SessionManager) VerifySession(sessionID string, context *RequestContext) (*Session, error) {
	if sm.sessionCache.IsActive(sessionID) {
		return sm.sessionCache.Get(sessionID), nil
	}

	if sm.fetchHandler == nil {
		return nil, nil
	}

	var session *Session
	var err error
	if context != nil {
		session, err = sm.fetchHandler(sessionID, context.Tx())
	} else {
		session, err = sm.fetchHandler(sessionID, nil)
	}
	if err != nil {
		return nil, err
	}
	if session == nil || session.IsZero() {
		return nil, nil
	}

	sm.sessionCache.Add(session)
	return session, nil
}

// Redirect returns a redirect result for when auth fails and you need to
// send the user to a login page.
func (sm *SessionManager) Redirect(context *RequestContext) ControllerResult {
	if sm.loginRedirectHandler != nil {
		redirectTo := context.auth.loginRedirectHandler(context.Request.URL)
		if redirectTo != nil {
			return context.Redirect(redirectTo.String())
		}
	}
	return context.DefaultResultProvider().NotAuthorized()
}
