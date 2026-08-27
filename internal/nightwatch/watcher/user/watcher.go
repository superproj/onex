// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

// Package user is a watcher implement.
package user

import (
	"context"

	"github.com/gammazero/workerpool"
	"github.com/looplab/fsm"
	"github.com/onexstack/onexstack/pkg/log"
	stringsutil "github.com/onexstack/onexstack/pkg/util/strings"
	"github.com/onexstack/onexstack/pkg/watch/registry"

	"github.com/onexstack/onex/internal/nightwatch/watcher"
	"github.com/onexstack/onex/internal/pkg/client/store"
	"github.com/onexstack/onex/internal/pkg/contextx"
	known "github.com/onexstack/onex/internal/pkg/known/usercenter"
	"github.com/onexstack/onex/internal/usercenter/model"
)

var _ registry.Watcher = (*Watcher)(nil)

// watcher implement.
type Watcher struct {
	store      store.Interface
	maxWorkers int64
}

// UserStateMachine is a struct that represents a user finite state machine.
type UserStateMachine struct {
	UserM *model.UserM
	FSM   *fsm.FSM
}

// Run runs the watcher.
func (w *Watcher) Run() {
	_, users, err := w.store.UserCenter().User().List(context.Background(), nil)
	if err != nil {
		log.Errorw(err, "Failed to list users")
		return
	}

	allowOperations := []string{
		// Need active user.
		known.UserStatusRegistered,
		// Need disable user.
		known.UserStatusBlacklisted,
		known.UserStatusNeedActive,
		known.UserStatusNeedDisable,
		// After disabling the user, they can be deleted, and the FSM will automatically transition to the next deleted state.
		// I have decided not to delete the user in the code, so the state transition here is commented out.
		// known.UserStatusDisabled,
	}

	wp := workerpool.New(int(w.maxWorkers))
	for _, user := range users {
		if !stringsutil.StringIn(user.Status, allowOperations) {
			continue
		}

		wp.Submit(func() {
			ctx := contextx.WithUserM(context.Background(), user)

			usm := &UserStateMachine{UserM: user, FSM: NewFSM(user.Status, w)}
			if err := usm.FSM.Event(ctx, user.Status); err != nil {
				log.Errorw(err, "Failed to event user", "username", user.Username, "status", user.Status)
				return
			}

			// When the entire state machine reaches the final state, print a message and send a notification.
			if usm.FSM.Current() == known.UserStatusDeleted {
				// We can add some lark card here in the future.
				log.Infow("Finish to handle user", "username", usm.UserM.Username)
			}

			return
		})
	}

	wp.StopWait()
}

// SetAggregateConfig initializes the watcher for later execution.
func (w *Watcher) SetAggregateConfig(config *watcher.AggregateConfig) {
	w.store = config.AggregateStore
	w.maxWorkers = config.UserWatcherMaxWorkers
}

func init() {
	registry.Register("user", &Watcher{})
}
