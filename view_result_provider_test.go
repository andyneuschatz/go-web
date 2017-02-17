package web

import (
	"bytes"
	"html/template"
	"net/http"
	"sync"
	"testing"

	"github.com/blendlabs/go-assert"
	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-logger"
)

func agent() *logger.DiagnosticsAgent {
	return logger.NewDiagnosticsAgent(logger.NewEventFlagSetNone())
}

func TestViewResultProviderNotFound(t *testing.T) {
	assert := assert.New(t)

	result := NewViewResultProvider(agent(), NewViewCache(), nil).NotFound()
	assert.NotNil(result)
	typed, isTyped := result.(*ViewResult)
	assert.True(isTyped)
	assert.Equal(http.StatusNotFound, typed.StatusCode)
}

func TestViewResultProviderNotAuthorized(t *testing.T) {
	assert := assert.New(t)

	result := NewViewResultProvider(agent(), NewViewCache(), nil).NotAuthorized()
	assert.NotNil(result)
	typed, isTyped := result.(*ViewResult)
	assert.True(isTyped)
	assert.Equal(http.StatusForbidden, typed.StatusCode)
}

func TestViewResultProviderInternalError(t *testing.T) {
	assert := assert.New(t)

	result := NewViewResultProvider(agent(), NewViewCache(), nil).InternalError(exception.New("Test"))
	assert.NotNil(result)
	typed, isTyped := result.(*ViewResult)
	assert.True(isTyped)
	assert.Equal(http.StatusInternalServerError, typed.StatusCode)
}

func TestViewResultProviderInternalErrorWritesToLogger(t *testing.T) {
	assert := assert.New(t)

	wg := sync.WaitGroup{}
	wg.Add(1)

	logBuffer := bytes.NewBuffer([]byte{})
	app := New()
	app.SetDiagnostics(logger.NewDiagnosticsAgent(logger.NewEventFlagSetWithEvents(logger.EventFatalError), logger.NewLogWriter(logBuffer)))
	app.diagnostics.AddEventListener(logger.EventFatalError, func(wr logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		defer wg.Done()
		assert.Len(state, 2)
		wr.Errorf("%v", state[0])
	})

	rc, err := app.Mock().Ctx(nil)
	assert.Nil(err)

	result := NewViewResultProvider(app.Diagnostics(), NewViewCache(), rc).InternalError(exception.New("Test"))
	assert.NotNil(result)
	typed, isTyped := result.(*ViewResult)
	assert.True(isTyped)
	assert.Equal(http.StatusInternalServerError, typed.StatusCode)

	wg.Wait()
	assert.NotZero(logBuffer.Len())
}

func TestViewResultProviderBadRequest(t *testing.T) {
	assert := assert.New(t)

	result := NewViewResultProvider(agent(), NewViewCache(), nil).BadRequest("test")
	assert.NotNil(result)
	typed, isTyped := result.(*ViewResult)
	assert.True(isTyped)
	assert.Equal(http.StatusBadRequest, typed.StatusCode)
}

type testViewModel struct {
	Text string
}

func TestViewResultProviderView(t *testing.T) {
	assert := assert.New(t)

	testView := template.New("testView")
	testView.Parse("{{.Text}}")

	provider := NewViewResultProvider(agent(), NewViewCache(), nil)
	provider.viewCache.SetTemplates(testView)
	result := provider.View("testView", testViewModel{Text: "foo"})

	assert.NotNil(result)
	typed, isTyped := result.(*ViewResult)
	assert.True(isTyped)
	assert.Equal(http.StatusOK, typed.StatusCode)
}
