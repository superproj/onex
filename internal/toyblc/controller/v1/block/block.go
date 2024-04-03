// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package block

import (
	"github.com/google/wire"

	"github.com/superproj/onex/internal/toyblc/blc"
	"github.com/superproj/onex/internal/toyblc/ws"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(New)

type BlockController struct {
	bs *blc.BlockSet
	ss *ws.Sockets
}

func New(bs *blc.BlockSet, ss *ws.Sockets) *BlockController {
	return &BlockController{bs, ss}
}
