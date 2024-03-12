// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package main

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/superproj/onex/pkg/log"
)

func main() {
	log.Infof("Start fake miner")
	for {
		data := []byte("hello, world!")
		hash := sha256.Sum256(data)
		fmt.Printf("%xn", hash)

		time.Sleep(30 * time.Second)
	}
}
