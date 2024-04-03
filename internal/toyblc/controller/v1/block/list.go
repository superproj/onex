// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package block

import (
	"github.com/gin-gonic/gin"

	"github.com/superproj/onex/internal/pkg/core"
)

func (b *BlockController) List(c *gin.Context) {
	core.WriteResponse(c, nil, b.bs.List())
}
