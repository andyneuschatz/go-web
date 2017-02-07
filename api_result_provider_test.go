package web

import (
	"bytes"
	"net/http"
	"sync"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-logger"
)

func TestAPIResultProviderNotFound(t *testing.T) {
	assert := assert.New(t)

	result := NewAPIResultProvider(agent(), nil).NotFound()
	assert.NotNil(result)
	typed, isTyped := result.(*JSONResult)
	assert.True(isTyped)
	assert.Equal(http.StatusNotFound, typed.StatusCode)
}

func TestAPIResultProviderNotAuthorized(t *testing.T) {
	assert := assert.New(t)

	result := NewAPIResultProvider(agent(), nil).NotAuthorized()
	assert.NotNil(result)
	typed, isTyped := result.(*JSONResult)
	assert.True(isTyped)
	assert.Equal(http.StatusForbidden, typed.StatusCode)
}

func TestAPIResultProviderInternalError(t *testing.T) {
	assert := assert.New(t)

	result := NewAPIResultProvider(agent(), nil).InternalError(exception.New("Test"))
	assert.NotNil(result)
	typed, isTyped := result.(*JSONResult)
	assert.True(isTyped)
	assert.Equal(http.StatusInternalServerError, typed.StatusCode)
}

func TestAPIResultProviderInternalErrorWritesToLogger(t *testing.T) {
	assert := assert.New(t)

	wg := sync.WaitGroup{}
	wg.Add(1)

	buffer := bytes.NewBuffer([]byte{})
	app := New()
	app.SetDiagnostics(logger.NewDiagnosticsAgent(logger.NewEventFlagSetWithEvents(logger.EventFatalError), logger.NewLogWriter(buffer)))
	app.diagnostics.AddEventListener(logger.EventFatalError, func(wr logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		defer wg.Done()
		wr.Errorf("%v", state[0])
		assert.Len(state, 2)
	})

	rc, err := app.Mock().RequestContext(nil)
	assert.Nil(err)
	result := rc.API().InternalError(exception.New("Test"))
	assert.NotNil(result)
	typed, isTyped := result.(*JSONResult)
	assert.True(isTyped)
	assert.Equal(http.StatusInternalServerError, typed.StatusCode)

	wg.Wait()
	assert.NotZero(buffer.Len())
}

func TestAPIResultProviderBadRequest(t *testing.T) {
	assert := assert.New(t)

	result := NewAPIResultProvider(agent(), nil).BadRequest("test")
	assert.NotNil(result)
	typed, isTyped := result.(*JSONResult)
	assert.True(isTyped)
	assert.Equal(http.StatusBadRequest, typed.StatusCode)
}

func TestAPIResultProviderOK(t *testing.T) {
	assert := assert.New(t)

	result := NewAPIResultProvider(agent(), nil).OK()
	assert.NotNil(result)
	typed, isTyped := result.(*JSONResult)
	assert.True(isTyped)
	assert.Equal(http.StatusOK, typed.StatusCode)
}

func TestAPIResultProviderJSON(t *testing.T) {
	assert := assert.New(t)

	result := NewAPIResultProvider(agent(), nil).JSON("foo")
	assert.NotNil(result)
	typed, isTyped := result.(*JSONResult)
	assert.True(isTyped)
	assert.Equal(http.StatusOK, typed.StatusCode)
}
