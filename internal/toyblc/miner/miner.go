// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package miner

import (
	"fmt"
	"time"

	"github.com/superproj/onex/internal/toyblc/blc"
	"github.com/superproj/onex/internal/toyblc/ws"
	"github.com/superproj/onex/pkg/log"
)

type Miner struct {
	bs              *blc.BlockSet
	ss              *ws.Sockets
	minMineInterval time.Duration
}

func NewMiner(bs *blc.BlockSet, ss *ws.Sockets, minMineInterval time.Duration) *Miner {
	return &Miner{bs: bs, ss: ss, minMineInterval: minMineInterval}
}

func (m *Miner) Start() {
	go func() {
		for {
			time.Sleep(interval(m.minMineInterval))
			block := MinerBlock(m.bs, m.ss, fmt.Sprintf("miner at %s", time.Now().Format("2006-01-02 15:04:05.000")))
			log.Debugw("Mine a block", "index", block.Index)
		}
	}()
}

func MinerBlock(bs *blc.BlockSet, ss *ws.Sockets, data string) *blc.Block {
	block := bs.NextBlock(data)
	bs.Add(block)
	ss.Broadcast(bs.LatestMessage())
	return block
}

func interval(minMineInterval time.Duration) time.Duration {
	return minMineInterval
}
