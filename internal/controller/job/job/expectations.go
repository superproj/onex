// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package job

import (
	"fmt"
	"sync/atomic"
	"time"

	"k8s.io/client-go/tools/cache"
)

// ExpectationsTimeout is how long unfulfilled expectations are kept before they
// are considered stale and no longer block reconciliation.
const ExpectationsTimeout = 5 * time.Minute

// ControlleeExpectations tracks the number of child writes the controller has
// issued but not yet observed through its informer. The counters are updated
// atomically because they are written both by the reconcile worker and by
// informer event handlers.
type ControlleeExpectations struct {
	add       int64
	del       int64
	key       string
	timestamp time.Time
}

// Fulfilled reports whether all expected creations and deletions have been observed.
func (e *ControlleeExpectations) Fulfilled() bool {
	return atomic.LoadInt64(&e.add) <= 0 && atomic.LoadInt64(&e.del) <= 0
}

func (e *ControlleeExpectations) GetExpectations() (int64, int64) {
	return atomic.LoadInt64(&e.add), atomic.LoadInt64(&e.del)
}

func (e *ControlleeExpectations) SetExpectations(add, del int) {
	atomic.StoreInt64(&e.add, int64(add))
	atomic.StoreInt64(&e.del, int64(del))
	e.timestamp = time.Now()
}

func (e *ControlleeExpectations) isExpired() bool {
	return time.Since(e.timestamp) > ExpectationsTimeout
}

func (e *ControlleeExpectations) addDelta(delta int) { atomic.AddInt64(&e.add, int64(delta)) }
func (e *ControlleeExpectations) delDelta(delta int) { atomic.AddInt64(&e.del, int64(delta)) }

// ControllerExpectations is a cache mapping a controller key (namespace/name) to
// its ControlleeExpectations.
type ControllerExpectations struct {
	cache.Store
}

// NewControllerExpectations returns a new ControllerExpectations cache.
func NewControllerExpectations() *ControllerExpectations {
	return &ControllerExpectations{cache.NewStore(expectationsKeyFunc)}
}

func expectationsKeyFunc(obj any) (string, error) {
	if e, ok := obj.(*ControlleeExpectations); ok {
		return e.key, nil
	}
	return "", fmt.Errorf("expected *ControlleeExpectations, got %T", obj)
}

// GetExpectations returns the expectations for the given key and whether they exist.
func (r *ControllerExpectations) GetExpectations(controllerKey string) (*ControlleeExpectations, bool, error) {
	exp, exists, err := r.GetByKey(controllerKey)
	if err == nil && exists {
		return exp.(*ControlleeExpectations), true, nil
	}
	return nil, false, err
}

// SatisfiedExpectations returns true if there are no pending expectations for
// the key, or if they have been fulfilled or expired.
func (r *ControllerExpectations) SatisfiedExpectations(controllerKey string) bool {
	if exp, exists, err := r.GetExpectations(controllerKey); exists {
		return exp.Fulfilled() || exp.isExpired()
	} else if err != nil {
		return false
	}
	return true
}

// DeleteExpectations drops the expectations for the given key.
func (r *ControllerExpectations) DeleteExpectations(controllerKey string) {
	exp, exists, err := r.GetExpectations(controllerKey)
	if err == nil && exists {
		_ = r.Delete(exp)
	}
}

// SetExpectations upserts the add/del counters for the given key.
func (r *ControllerExpectations) SetExpectations(controllerKey string, add, del int) {
	if exp, exists, _ := r.GetExpectations(controllerKey); exists {
		exp.SetExpectations(add, del)
		_ = r.Update(exp)
		return
	}
	exp := &ControlleeExpectations{key: controllerKey, timestamp: time.Now()}
	exp.SetExpectations(add, del)
	_ = r.Add(exp)
}

// ExpectCreations increments the expected-creations counter for the key.
func (r *ControllerExpectations) ExpectCreations(controllerKey string, adds int) {
	e := r.getOrCreate(controllerKey)
	e.addDelta(adds)
	_ = r.Update(e)
}

// CreationObserved decrements the expected-creations counter for the key.
func (r *ControllerExpectations) CreationObserved(controllerKey string) {
	e := r.getOrCreate(controllerKey)
	e.addDelta(-1)
	_ = r.Update(e)
}

// ExpectDeletions increments the expected-deletions counter for the key.
func (r *ControllerExpectations) ExpectDeletions(controllerKey string, dels int) {
	e := r.getOrCreate(controllerKey)
	e.delDelta(dels)
	_ = r.Update(e)
}

// DeletionObserved decrements the expected-deletions counter for the key.
func (r *ControllerExpectations) DeletionObserved(controllerKey string) {
	e := r.getOrCreate(controllerKey)
	e.delDelta(-1)
	_ = r.Update(e)
}

func (r *ControllerExpectations) getOrCreate(controllerKey string) *ControlleeExpectations {
	if exp, exists, _ := r.GetExpectations(controllerKey); exists {
		return exp
	}
	exp := &ControlleeExpectations{key: controllerKey, timestamp: time.Now()}
	_ = r.Add(exp)
	return exp
}
