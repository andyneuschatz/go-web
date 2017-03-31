package web

import logger "github.com/blendlabs/go-logger"

const (
	// EventWebRequestStart is an aliased event flag.
	EventWebRequestStart = logger.EventWebRequestStart

	// EventWebRequest is an aliased event flag.
	EventWebRequest = logger.EventWebRequest

	// EventWebResponse is an aliased event flag.
	EventWebResponse = logger.EventWebResponse

	// EventWebRequestPostBody is an aliased event flag.
	EventWebRequestPostBody = logger.EventWebRequestPostBody
)

// RequestListener is a listener for `EventRequestStart` and `EventRequest` events.
type RequestListener func(logger.Logger, logger.TimeSource, *Ctx)

// NewRequestListener creates a new logger.EventListener for `EventRequestStart` and `EventRequest` events.
func NewRequestListener(listener RequestListener) logger.EventListener {
	return func(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		listener(writer, ts, state[0].(*Ctx))
	}
}

// ErrorListener is a listener for errors with an associated request context.
type ErrorListener func(logger.Logger, logger.TimeSource, error, *Ctx)

// NewErrorListener returns a new error listener.
func NewErrorListener(listener ErrorListener) logger.EventListener {
	return func(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
		err := state[0].(error)

		if len(state) > 1 {
			if ctx, hasCtx := state[1].(*Ctx); hasCtx {
				listener(writer, ts, err, ctx)
			}
		}
		listener(writer, ts, err, nil)
	}
}
