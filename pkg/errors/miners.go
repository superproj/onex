// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package errors

import (
	"fmt"
)

// MinerError is a more descriptive kind of error that represents an error condition that
// should be set in the Miner.Status. The "Reason" field is meant for short,
// enum-style constants meant to be interpreted by miners. The "Message"
// field is meant to be read by humans.
type MinerError struct {
	Reason  MinerStatusError
	Message string
}

func (e *MinerError) Error() string {
	return e.Message
}

// Some error builders for ease of use. They set the appropriate "Reason"
// value, and all arguments are Printf-style varargs fed into Sprintf to
// construct the Message.

// InvalidMinerConfiguration creates a new error when a Miner has invalid configuration.
func InvalidMinerConfiguration(msg string, args ...any) *MinerError {
	return &MinerError{
		Reason:  InvalidConfigurationMinerError,
		Message: fmt.Sprintf(msg, args...),
	}
}

// CreateMiner creates a new error for when creating a Miner.
func CreateMiner(msg string, args ...any) *MinerError {
	return &MinerError{
		Reason:  CreateMinerError,
		Message: fmt.Sprintf(msg, args...),
	}
}

// UpdateMiner creates a new error for when updating a Miner.
func UpdateMiner(msg string, args ...any) *MinerError {
	return &MinerError{
		Reason:  UpdateMinerError,
		Message: fmt.Sprintf(msg, args...),
	}
}

// DeleteMiner creates a new error for when deleting a Miner.
func DeleteMiner(msg string, args ...any) *MinerError {
	return &MinerError{
		Reason:  DeleteMinerError,
		Message: fmt.Sprintf(msg, args...),
	}
}
