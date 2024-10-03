package fsm

import (
	"context"

	"github.com/looplab/fsm"
)

// WrapEvent wraps a function that handles FSM events, allowing it to return an error.
func WrapEvent(fn func(ctx context.Context, event *fsm.Event) error) fsm.Callback {
	return func(ctx context.Context, event *fsm.Event) {
		if err := fn(ctx, event); err != nil {
			event.Err = err
		}
	}
}
