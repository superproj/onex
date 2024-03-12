// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package id

import (
	"context"
	"fmt"
	"time"

	"github.com/sony/sonyflake"
)

type Sonyflake struct {
	ops   SonyflakeOptions
	sf    *sonyflake.Sonyflake
	Error error
}

// NewSonyflake can get a unique code by id(You need to ensure that id is unique).
func NewSonyflake(options ...func(*SonyflakeOptions)) *Sonyflake {
	ops := getSonyflakeOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	sf := &Sonyflake{
		ops: *ops,
	}
	st := sonyflake.Settings{
		StartTime: ops.startTime,
	}
	if ops.machineId > 0 {
		st.MachineID = func() (uint16, error) {
			return ops.machineId, nil
		}
	}
	ins := sonyflake.NewSonyflake(st)
	if ins == nil {
		sf.Error = fmt.Errorf("create snoyflake failed")
	}
	_, err := ins.NextID()
	if err != nil {
		sf.Error = fmt.Errorf("invalid start time")
	}
	sf.sf = ins
	return sf
}

func (s *Sonyflake) Id(ctx context.Context) (id uint64) {
	if s.Error != nil {
		return
	}
	var err error
	id, err = s.sf.NextID()
	if err == nil {
		return
	}

	sleep := 1
	for {
		time.Sleep(time.Duration(sleep) * time.Millisecond)
		id, err = s.sf.NextID()
		if err == nil {
			return
		}
		sleep *= 2
	}
}
