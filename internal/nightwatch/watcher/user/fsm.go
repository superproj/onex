package user

import (
	"github.com/looplab/fsm"

	known "github.com/superproj/onex/internal/pkg/known/usercenter"
)

// NewFSM creates a new finite state machine (FSM) for managing user states.
// The function takes an initial user status and a user watcher as input parameters.
// The FSM is configured with the following events and callbacks:
//
// Events:
// - UserStatusRegistered -> UserStatusActived
// - UserStatusBlacklisted -> UserStatusDisabled
// - UserStatusDisabled -> UserStatusDeleted
//
// Callbacks:
// - UserStatusActived: Calls the NewActiveUserCallback function to handle the "active user" event.
// - UserStatusDisabled: Calls the NewDisableUserCallback function to handle the "disable user" event.
// - UserStatusDeleted: Calls the NewDeleteUserCallback function to handle the "delete user" event.
// - UserEventAfterEvent: Calls the NewUserEventAfterEvent function after any user-related event is handled.
//
// The function returns the newly created FSM.
func NewFSM(initial string, w *userWatcher) *fsm.FSM {
	return fsm.NewFSM(
		initial,
		fsm.Events{
			// Define status events.
			{Name: known.UserStatusRegistered, Src: []string{known.UserStatusRegistered}, Dst: known.UserStatusActived},
			{Name: known.UserStatusBlacklisted, Src: []string{known.UserStatusBlacklisted}, Dst: known.UserStatusDisabled},
			// Define need events.
			{Name: known.UserStatusNeedActive, Src: []string{known.UserStatusNeedActive}, Dst: known.UserStatusActived},
			{Name: known.UserStatusNeedDisable, Src: []string{known.UserStatusNeedDisable}, Dst: known.UserStatusDisabled},
			// After disabling the user, they can be deleted, and the FSM will automatically transition to the next deleted state.
			// I have decided not to delete the user in the code, so the state transition here is commented out.
			// {Name: known.UserStatusDisabled, Src: []string{known.UserStatusDisabled}, Dst: known.UserStatusDeleted},
		},
		fsm.Callbacks{
			known.UserStatusActived:   NewActiveUserCallback(w.store),
			known.UserStatusDisabled: NewDisableUserCallback(w.store),
			known.UserStatusDeleted:  NewDeleteUserCallback(w.store),
			// log, alert, save to stoer, etc for all events.
			// Alert the status of each step of the operation.
			UserEventAfterEvent: NewUserEventAfterEvent(w.store),
		},
	)
}
