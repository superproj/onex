// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package job

import (
	"sync"
	"time"
)

const (
	// defaultBackoff is the base delay applied after the first failed dispatch.
	defaultBackoff = 5 * time.Second
	// maxBackoff caps the exponential backoff between retries.
	maxBackoff = 5 * time.Minute
)

// backoffRecord tracks the retry state of a single Job whose dispatch has failed.
// It is the onex port of k8s.io/kubernetes/pkg/controller/job's backoffRecord.
type backoffRecord struct {
	// failuresAfterLastSuccess counts consecutive dispatch failures.
	failuresAfterLastSuccess int
	// lastFailureTime is the time of the most recent failure.
	lastFailureTime time.Time
}

// getRemainingTime returns how long to wait before retrying, applying
// exponential decay (defaultBackoff * 2^n, capped at maxBackoff) minus the time
// already elapsed since the last failure.
func (b backoffRecord) getRemainingTime(now time.Time, baseDelay, maxDelay time.Duration) time.Duration {
	if b.failuresAfterLastSuccess <= 0 {
		return 0
	}

	backoff := baseDelay
	for i := 1; i < b.failuresAfterLastSuccess; i++ {
		backoff *= 2
		if backoff > maxDelay {
			backoff = maxDelay
			break
		}
	}
	if backoff > maxDelay {
		backoff = maxDelay
	}

	remaining := backoff - now.Sub(b.lastFailureTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// backoffStore keeps per-Job backoff records. It is keyed by the Job's
// namespace/name.
type backoffStore struct {
	mu      sync.Mutex
	records map[string]backoffRecord
}

func newBackoffStore() *backoffStore {
	return &backoffStore{records: map[string]backoffRecord{}}
}

// getRemainingTime returns the remaining backoff for the key, or 0.
func (s *backoffStore) getRemainingTime(key string, now time.Time) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rec, ok := s.records[key]; ok {
		return rec.getRemainingTime(now, defaultBackoff, maxBackoff)
	}
	return 0
}

// recordFailure increments the failure count and stamps the failure time.
func (s *backoffStore) recordFailure(key string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec := s.records[key]
	rec.failuresAfterLastSuccess++
	rec.lastFailureTime = now
	s.records[key] = rec
}

// reset clears the backoff record for the key (e.g. after a successful dispatch).
func (s *backoffStore) reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, key)
}
