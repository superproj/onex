// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package nightwatch

import (
	"github.com/superproj/onex/pkg/log"
)

// cronLogger implement the cron.Logger interface.
type cronLogger struct{}

// newCronLogger returns a cron logger.
func newCronLogger() *cronLogger {
	return &cronLogger{}
}

// Info logs routine messages about cron's operation.
func (l *cronLogger) Info(msg string, keysAndValues ...any) {
	log.Infow(msg, keysAndValues...)
}

// Error logs an error condition.
func (l *cronLogger) Error(err error, msg string, keysAndValues ...any) {
	log.Errorw(err, msg, keysAndValues...)
}
