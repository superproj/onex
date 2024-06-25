package user

import (
	"context"
	"time"

	"github.com/looplab/fsm"

	"github.com/superproj/onex/internal/pkg/client/store"
	known "github.com/superproj/onex/internal/pkg/known/usercenter"
	"github.com/superproj/onex/internal/pkg/onexx"
	"github.com/superproj/onex/pkg/log"
)

const (
	UserEventAfterEvent = "after_event"
)

// NewActiveUserCallback creates a callback function for the "active user" event in a finite state machine (FSM).
func NewActiveUserCallback(store store.Interface) fsm.Callback {
	return func(ctx context.Context, event *fsm.Event) {
		userM := onexx.FromUserM(ctx)
		log.Infow("Now active user", "event", event.Event, "username", userM.Username)
		// Fake active user operations.
		time.Sleep(5 * time.Second)
		log.Infow("Success to active user", "event", event.Event, "username", userM.Username)
	}
}

// NewDisableUserCallback creates a callback function for the "disable user" event in a finite state machine (FSM).
func NewDisableUserCallback(store store.Interface) fsm.Callback {
	return func(ctx context.Context, event *fsm.Event) {
		userM := onexx.FromUserM(ctx)
		log.Infow("Now disable user", "event", event.Event, "username", userM.Username)
		// Fake disable user operations.
		time.Sleep(5 * time.Second)
		log.Infow("Success to disable user", "event", event.Event, "username", userM.Username)
	}
}

// NewDeleteUserCallback creates a callback function for the "delete user" event in a finite state machine (FSM).
func NewDeleteUserCallback(store store.Interface) fsm.Callback {
	return func(ctx context.Context, event *fsm.Event) {
		userM := onexx.FromUserM(ctx)
		log.Infow("Now delete user", "event", event.Event, "username", userM.Username)
		// Fake delete user operations.
		time.Sleep(5 * time.Second)
		log.Infow("Success to delete user", "event", event.Event, "username", userM.Username)
	}
}

// NewUserEventAfterEvent creates a callback function that is executed after a
// user-related event is handled in a finite state machine (FSM).
func NewUserEventAfterEvent(store store.Interface) fsm.Callback {
	return func(ctx context.Context, event *fsm.Event) {
		alarmStatus := "success"
		userM := onexx.FromUserM(ctx)

		defer func() {
			log.Infow("This is a fake alarm message", "status", alarmStatus, "username", userM.Username)
		}()

		if event.Err != nil {
			alarmStatus = "failed"
			log.Errorw(event.Err, "Failed to handle event", "event", event.Event)
			// We can add some alerts here in the future.
			return
		}

		user := onexx.FromUserM(ctx)
		user.Status = event.FSM.Current()
		if err := store.UserCenter().Users().Update(ctx, user); err != nil {
			log.Errorw(err, "Failed to update status into database", "event", event.Event)
		}

		if user.Status == known.UserStatusDeleted {
			// We can add some lark card here in the future.
			log.Infow("Finish to handle user", "event", event.Event, "username", user.Username)
		}
	}
}
