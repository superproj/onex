// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package meta

const (
	// ListAll is the default argument to specify on a context when you want to list or filter resources across all scopes.
	ListAll      = ""
	defaultLimit = 1000
)

type ListOption func(*ListOptions)

type ListOptions struct {
	// Filters specify the equality where conditions.
	Filters map[string]any
	Offset  int
	Limit   int
}

func NewListOptions(opts ...ListOption) ListOptions {
	los := ListOptions{
		Filters: map[string]any{},
		Offset:  0,
		Limit:   defaultLimit,
	}

	for _, opt := range opts {
		opt(&los)
	}

	return los
}

func WithFilter(filter map[string]any) ListOption {
	return func(o *ListOptions) {
		o.Filters = filter
	}
}

func WithOffset(offset int64) ListOption {
	return func(o *ListOptions) {
		if offset < 0 {
			offset = 0
		}

		o.Offset = int(offset)
	}
}

func WithLimit(limit int64) ListOption {
	return func(o *ListOptions) {
		if limit <= 0 {
			limit = defaultLimit
		}

		o.Limit = int(limit)
	}
}
