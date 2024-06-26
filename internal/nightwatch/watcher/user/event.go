package user

import (
	"context"
	"time"

	"github.com/looplab/fsm"

	"github.com/superproj/onex/internal/pkg/client/store"
	known "github.com/superproj/onex/internal/pkg/known/usercenter"
	"github.com/superproj/onex/internal/pkg/onexx"
	"github.com/superproj/onex/internal/usercenter/model"
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

		// Active secrets if needed.
		if err := iterateSecrets(ctx, store, userM.UserID, activeSecret); err != nil {
			event.Err = err
			return
		}

		log.Infow("Success to active user", "event", event.Event, "username", userM.Username)
	}
}

// NewDisableUserCallback creates a callback function for the "disable user" event in a finite state machine (FSM).
func NewDisableUserCallback(store store.Interface) fsm.Callback {
	return func(ctx context.Context, event *fsm.Event) {
		userM := onexx.FromUserM(ctx)
		log.Infow("Now disable user", "event", event.Event, "username", userM.Username)

		// Disable secrets if needed.
		if err := iterateSecrets(ctx, store, userM.UserID, disableSecret); err != nil {
			event.Err = err
			return
		}

		log.Infow("Success to disable user", "event", event.Event, "username", userM.Username)
	}
}

// NewDeleteUserCallback creates a callback function for the "delete user" event in a finite state machine (FSM).
func NewDeleteUserCallback(store store.Interface) fsm.Callback {
	return func(ctx context.Context, event *fsm.Event) {
		userM := onexx.FromUserM(ctx)
		log.Infow("Now delete user if needed", "event", event.Event, "username", userM.Username)

		// If a user remains in an disalbed state for more than 5 years,
		// the user should be deleted.
		duration := time.Since(userM.UpdatedAt)
		if duration.Hours() < 24*365*5 {
			return
		}

		// Delete secrets if needed.
		if err := iterateSecrets(ctx, store, userM.UserID, deleteSecret); err != nil {
			event.Err = err
			return
		}

		// Save user data for archiving purposes.
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

// activeSecret used to active user secret.
func activeSecret(ctx context.Context, store store.Interface, secret *model.SecretM) error {
	log.Infow("Now actice user secret", "userID", secret.UserID, "secretID", secret.SecretID)
	// To avoid unnecessary database update operations, we first check
	// whether updating the database is required.
	if secret.Status == known.SecretStatusNormal {
		return nil
	}
	secret.Status = known.SecretStatusNormal
	return store.UserCenter().Secrets().Update(ctx, secret)
}

// disableSecret used to disable user secret.
func disableSecret(ctx context.Context, store store.Interface, secret *model.SecretM) error {
	log.Infow("Now disable user secret", "userID", secret.UserID, "secretID", secret.SecretID)
	// To avoid unnecessary database update operations, we first check
	// whether updating the database is required.
	if secret.Status == known.SecretStatusDisabled {
		return nil
	}
	secret.Status = known.SecretStatusDisabled
	return store.UserCenter().Secrets().Update(ctx, secret)
}

// deleteSecret used to delete user secret.
func deleteSecret(ctx context.Context, store store.Interface, secret *model.SecretM) error {
	log.Infow("Now delete user secret", "userID", secret.UserID, "secretID", secret.SecretID)
	return store.UserCenter().Secrets().Delete(ctx, secret.UserID, secret.Name)
}

// iterateSecrets iterates through the secrets of a user specified by userID
// and calls the action function on each secret.
func iterateSecrets(
	ctx context.Context,
	store store.Interface,
	userID string,
	action func(ctx context.Context, store store.Interface, secret *model.SecretM) error,
) error {
	// Retrieve the list of secrets for the specified user.
	_, secrets, err := store.UserCenter().Secrets().List(ctx, userID)
	if err != nil {
		return err
	}

	// Iterate through each secret and perform the action function.
	for i := range secrets {
		if err := action(ctx, store, secrets[i]); err != nil {
			return err
		}
	}

	return nil
}
