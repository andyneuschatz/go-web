package web

import (
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestNestMiddleware(t *testing.T) {
	assert := assert.New(t)

	var callIndex int

	var mw1Called int
	mw1 := func(action ControllerAction) ControllerAction {
		return func(rc *RequestContext) ControllerResult {
			mw1Called = callIndex
			callIndex = callIndex + 1
			return action(rc)
		}
	}

	var mw2Called int
	mw2 := func(action ControllerAction) ControllerAction {
		return func(rc *RequestContext) ControllerResult {
			mw2Called = callIndex
			callIndex = callIndex + 1
			return action(rc)
		}
	}

	var mw3Called int
	mw3 := func(action ControllerAction) ControllerAction {
		return func(rc *RequestContext) ControllerResult {
			mw3Called = callIndex
			callIndex = callIndex + 1
			return action(rc)
		}
	}

	nested := NestMiddleware(func(rc *RequestContext) ControllerResult { return nil }, mw2, mw3, mw1)

	nested(nil)

	assert.Equal(2, mw2Called)
	assert.Equal(1, mw3Called)
	assert.Equal(0, mw1Called)
}
