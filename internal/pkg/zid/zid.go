// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package zid

import (
	"github.com/superproj/onex/pkg/id"
)

const defaultABC = "abcdefghijklmnopqrstuvwxyz1234567890"

type ZID string

const (
	// ID for the user resource in onex-usercenter.
	User ZID = "user"
	// ID for the order resource in onex-fakeserver.
	Order ZID = "order"
	// ID for the cronjob resource in onex-nightwatch.
	CronJob ZID = "cronjob"
	// ID for the job resource in onex-nightwatch.
	Job ZID = "job"
)

func (zid ZID) String() string {
	return string(zid)
}

func (zid ZID) New(i uint64) string {
	// use custom option
	str := id.NewCode(
		i,
		id.WithCodeChars([]rune(defaultABC)),
		id.WithCodeL(6),
		id.WithCodeSalt(Salt()),
	)
	return zid.String() + "-" + str
}
