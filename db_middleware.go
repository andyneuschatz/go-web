package web

import "github.com/blendlabs/spiffy"

// WithDefaultDB sets the default database connection for a context.
func WithDefaultDB(action Action) Action {
	return func(context *Ctx) Result {
		if context.DB() == nil { //preserve testing db contexts
			return action(context.WithDB(spiffy.Default().DB()))
		}
		return action(context)
	}
}
