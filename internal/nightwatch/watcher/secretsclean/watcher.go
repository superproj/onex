// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package secretsclean is a watcher implement used to delete expired keys from the database.
package secretsclean

import (
	"context"
	"time"

	"github.com/superproj/onex/internal/nightwatch/watcher"
	"github.com/superproj/onex/internal/pkg/client/store"
	"github.com/superproj/onex/pkg/log"
)

var _ watcher.Watcher = (*secretsCleanWatcher)(nil)

// watcher implement.
type secretsCleanWatcher struct {
	store store.Interface
}

// Run runs the watcher.
func (w *secretsCleanWatcher) Run() {
	_, secrets, err := w.store.UserCenter().Secrets().List(context.Background(), "")
	if err != nil {
		log.Errorw(err, "Failed to list secrets")
		return
	}

	for _, secret := range secrets {
		if secret.Expires != 0 && secret.Expires < time.Now().AddDate(0, 0, -7).Unix() {
			err := w.store.UserCenter().Secrets().Delete(context.TODO(), secret.UserID, secret.Name)
			if err != nil {
				log.Warnw("Failed to delete secret from database", "userID", secret.UserID, "name", secret.Name)
				continue
			}
			log.Infow("Successfully deleted secret from database", "userID", secret.UserID, "name", secret.Name)
		}
	}
}

// Init initializes the watcher for later execution.
func (w *secretsCleanWatcher) Init(ctx context.Context, config *watcher.Config) error {
	w.store = config.Store
	return nil
}

func init() {
	watcher.Register("secretsclean", &secretsCleanWatcher{})
}
