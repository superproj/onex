// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package auth

import (
	"time"

	"github.com/casbin/casbin/v2"
	clog "github.com/casbin/casbin/v2/log"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	rediswatcher "github.com/casbin/redis-watcher/v2"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/superproj/onex/pkg/log"
	genericoptions "github.com/superproj/onex/pkg/options"
)

const (
	rbacModel = `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act, eft

[role_definition]
g = _, _

[policy_effect]
# The Effect primitive indicates that if there is no matching rule whose decision
# result is deny, and the final decision result is allow, that is, deny-override
# More effect syntax reference: https://casbin.org/docs/syntax-for-models#policy-effect
e = !some(where (p.eft == deny))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act || r.sub == "root"`
)

// AuthzProviderSet defines a wire set for authorization.
var AuthzProviderSet = wire.NewSet(NewAuthz, wire.Bind(new(AuthzInterface), new(*authzImpl)), LoggerProviderSet)

// AuthzInterface defines the interface for authorization.
type AuthzInterface interface {
	Authorize(rvals ...any) (bool, error)
}

type authzImpl struct {
	enforcer *casbin.SyncedEnforcer
}

// Ensure authzImpl implements AuthzInterface.
var _ AuthzInterface = (*authzImpl)(nil)

// updateCallback defines the function to be called when a policy update is detected.
func updateCallback(rev string) {
	log.Warnw("New revision detected", "revision", rev)
}

// NewAuthz creates a new authorization instance using the provided database, Redis options, and logger.
func NewAuthz(db *gorm.DB, redisOpts *genericoptions.RedisOptions, logger clog.Logger) (*authzImpl, error) {
	// Initialize a Gorm adapter and use it in a Casbin enforcer
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		log.Errorw(err, "Failed to initialize casbin adapter")
		return nil, err
	}

	// Initialize the watcher using Redis as a backend.
	w, err := rediswatcher.NewWatcher(redisOpts.Addr, rediswatcher.WatcherOptions{
		Options: redis.Options{
			DB:       redisOpts.Database,
			Username: redisOpts.Username,
			Password: redisOpts.Password,
		},
		Channel: "/casbin",
	})
	if err != nil {
		log.Errorw(err, "Failed to create casbin watcher")
		return nil, err
	}

	m, _ := model.NewModelFromString(rbacModel)

	// Initialize the enforcer.
	enforcer, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		log.Errorw(err, "Failed to create casbin enforcer")
		return nil, err
	}

	// Set the watcher for the enforcer.
	_ = enforcer.SetWatcher(w)
	// Change current casbin's logger to kafka logger
	enforcer.SetLogger(logger)
	// Enable casbin log messages to the kafka logger.
	enforcer.EnableLog(true)

	// By default, the watcher's callback is automatically set to the
	// enforcer's LoadPolicy() in the SetWatcher() call.
	// We can change it by explicitly setting a callback.
	_ = w.SetUpdateCallback(updateCallback)

	// Load the policy from DB.
	if err := enforcer.LoadPolicy(); err != nil {
		log.Errorw(err, "Failed to load casbin policy")
		return nil, err
	}
	// Start auto-loading the policy every minute.
	enforcer.StartAutoLoadPolicy(time.Minute)

	// Create a new authzImpl instance.
	return &authzImpl{enforcer: enforcer}, nil
}

// Authorize checks whether the given request values satisfy the authorization policy.
func (a *authzImpl) Authorize(rvals ...any) (bool, error) {
	return a.enforcer.Enforce(rvals...)
}
