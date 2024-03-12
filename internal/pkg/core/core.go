// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package core

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/errors"
)

func WriteResponse(c *gin.Context, err error, data any) {
	if err != nil {
		c.JSON(errors.Code(err), err)
		return
	}

	c.JSON(http.StatusOK, data)
}
