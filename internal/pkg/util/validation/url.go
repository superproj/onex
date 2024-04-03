// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package validation

import (
	"fmt"
	"regexp"
)

var urlRegex = regexp.MustCompile("(https?|ftp|file)://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]")

// IsURL test whether the given value is a valid URL address.
func IsURL(value string) error {
	if !urlRegex.MatchString(value) {
		return fmt.Errorf("not a valid URL address")
	}
	return nil
}
