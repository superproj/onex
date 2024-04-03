// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package streams

// Inlet represents a type that exposes one open input.
type Inlet interface {
	In() chan<- any
}

// Outlet represents a type that exposes one open output.
type Outlet interface {
	Out() <-chan any
}

// Source represents a set of stream processing steps that has one open output.
type Source interface {
	Outlet
	Via(Flow) Flow
}

// Flow represents a set of stream processing steps that has one open input and one open output.
type Flow interface {
	Inlet
	Outlet
	Via(Flow) Flow
	To(Sink)
}

// Sink represents a set of stream processing steps that has one open input.
type Sink interface {
	Inlet
}
