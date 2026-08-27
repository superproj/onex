// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

// Package secretsclean is a watcher implement used to delete expired keys from the database.
package secretsclean

import (
	"context"
	"time"

	"github.com/onexstack/onexstack/pkg/log"
	"github.com/onexstack/onexstack/pkg/store/where"
	"github.com/onexstack/onexstack/pkg/watch/registry"

	"github.com/onexstack/onex/internal/nightwatch/watcher"
	"github.com/onexstack/onex/internal/pkg/client/store"
)

var _ registry.Watcher = (*Watcher)(nil)

// watcher implement.
type Watcher struct {
	store store.Interface
}

// Run runs the watcher.
func (w *Watcher) Run() {
	ctx := context.Background()
	_, secrets, err := w.store.UserCenter().Secret().List(ctx, nil)
	if err != nil {
		log.Errorw(err, "Failed to list secrets")
		return
	}

	for _, secret := range secrets {
		if secret.Expires != 0 && secret.Expires < time.Now().AddDate(0, 0, -7).Unix() {
			err := w.store.UserCenter().Secret().Delete(ctx, where.F("user_id", secret.UserID, "name", secret.Name))
			if err != nil {
				log.Warnw("Failed to delete secret from database", "userID", secret.UserID, "name", secret.Name)
				continue
			}
			log.Infow("Successfully deleted secret from database", "userID", secret.UserID, "name", secret.Name)
		}
	}
}

// SetAggregateConfig initializes the watcher for later execution.
func (w *Watcher) SetAggregateConfig(config *watcher.AggregateConfig) {
	w.store = config.AggregateStore
}

func init() {
	registry.Register("secretsclean", &Watcher{})
}
