// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package i18n

import (
	"fmt"
	"testing"

	"golang.org/x/text/language"
)

// //go:embed locales
// var fs embed.FS

func TestNew(t *testing.T) {
	i := New()
	// 1. add dir
	i.Add("./locales")

	// 2. add file
	// i.Add("./locales/en.yml")
	// i.Add("./locales/zh.yml")

	// 3. add embed fs
	// i.AddFs(fs)

	fmt.Println(i.T("common.hello"))
	fmt.Println(i.Select(language.Chinese).T("common.hello"))
}
