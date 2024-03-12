// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package auth

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	clog "github.com/casbin/casbin/v2/log"
	"github.com/google/wire"
	"github.com/segmentio/kafka-go"

	"github.com/superproj/onex/pkg/log"
	genericoptions "github.com/superproj/onex/pkg/options"
)

// LoggerProviderSet defines a wire set for creating a kafkaLogger instance to implement log.Logger interface.
var LoggerProviderSet = wire.NewSet(NewLogger, wire.Bind(new(clog.Logger), new(*kafkaLogger)))

// kafkaLogger is a log.Logger implementation that writes log messages to Kafka.
type kafkaLogger struct {
	// enabled is an atomic boolean indicating whether the logger is enabled.
	enabled int32
	// writer is the Kafka writer used to write log messages.
	writer *kafka.Writer
}

// AuditMessage is the message structure for log messages.
type AuditMessage struct {
	Matcher   string        `protobuf:"bytes,1,opt,name=matcher,proto3" json:"matcher,omitempty"`
	Request   []any `protobuf:"bytes,2,opt,name=request,proto3" json:"request,omitempty"`
	Result    bool          `protobuf:"bytes,3,opt,name=result,proto3" json:"result,omitempty"`
	Explains  [][]string    `protobuf:"bytes,4,opt,name=explains,proto3" json:"explains,omitempty"`
	Timestamp int64         `protobuf:"bytes,5,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

// NewLogger creates a new kafkaLogger instance.
func NewLogger(kafkaOpts *genericoptions.KafkaOptions) (*kafkaLogger, error) {
	writer, err := kafkaOpts.Writer()
	if err != nil {
		return nil, err
	}

	return &kafkaLogger{writer: writer}, nil
}

// EnableLog enables or disables the logger.
func (l *kafkaLogger) EnableLog(enable bool) {
	var enab int32
	if enable {
		enab = 1
	}
	atomic.StoreInt32(&l.enabled, enab)
}

// IsEnabled returns whether the logger is enabled.
func (l *kafkaLogger) IsEnabled() bool {
	return atomic.LoadInt32(&l.enabled) == 1
}

// LogEnforce writes a log message for a policy enforcement decision.
func (l *kafkaLogger) LogModel(model [][]string) {
	if !l.IsEnabled() {
		return
	}
	log.Debugw("LogModel", "model", model)
}

// LogModel writes a log message for the policy model.
func (l *kafkaLogger) LogEnforce(matcher string, request []any, result bool, explains [][]string) {
	if !l.IsEnabled() {
		return
	}

	message := AuditMessage{
		Matcher:   matcher,
		Request:   request,
		Result:    result,
		Explains:  explains,
		Timestamp: time.Now().Unix(),
	}

	out, _ := json.Marshal(message)
	if err := l.writer.WriteMessages(context.Background(), kafka.Message{Value: out}); err != nil {
		log.Errorw(err, "Failed to write kafka messages")
	}
	log.Debugw("LogEnforce", "matcher", matcher, "request", request, "result", result, "explains", explains)
}

// LogPolicy writes a log message for the policy rules.
func (l *kafkaLogger) LogPolicy(policy map[string][][]string) {
	if !l.IsEnabled() {
		return
	}
	log.Debugw("LogPolicy", "policy", policy)
}

// LogRole writes a log message for the policy roles.
func (l *kafkaLogger) LogRole(roles []string) {
	if !l.IsEnabled() {
		return
	}
	log.Debugw("LogRole", "roles", roles)
}
