// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

// Package all blank-imports every Job provider so that their init functions
// register themselves with the provider registry. Adding a new provider is a
// matter of appending its package to the import list below.
package all

import (
	_ "github.com/onexstack/onex/internal/controller/job/job/providers/kubernetes"
	_ "github.com/onexstack/onex/internal/controller/job/job/providers/llmtrain"
)
