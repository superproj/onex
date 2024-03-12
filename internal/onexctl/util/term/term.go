// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package term provides structures and helper functions to work with
// terminal (state, sizes).
package term

import (
	"io"
	"os"
	"runtime"

	"github.com/moby/term"
)

// TTY helps invoke a function and preserve the state of the terminal, even if the process is
// terminated during execution. It also provides support for terminal resizing for remote command
// execution/attachment.
type TTY struct {
	// In is a reader representing stdin. It is a required field.
	In io.Reader
	// Out is a writer representing stdout. It must be set to support terminal resizing. It is an
	// optional field.
	Out io.Writer
	// Raw is true if the terminal should be set raw.
	Raw bool
	// TryDev indicates the TTY should try to open /dev/tty if the provided input
	// is not a file descriptor.
	TryDev bool
}

// IsTerminalIn returns true if t.In is a terminal. Does not check /dev/tty
// even if TryDev is set.
func (t TTY) IsTerminalIn() bool {
	return IsTerminal(t.In)
}

// IsTerminalOut returns true if t.Out is a terminal. Does not check /dev/tty
// even if TryDev is set.
func (t TTY) IsTerminalOut() bool {
	return IsTerminal(t.Out)
}

// IsTerminal returns whether the passed object is a terminal or not.
func IsTerminal(i any) bool {
	_, terminal := term.GetFdInfo(i)
	return terminal
}

// AllowsColorOutput returns true if the specified writer is a terminal and
// the process environment indicates color output is supported and desired.
func AllowsColorOutput(w io.Writer) bool {
	if !IsTerminal(w) {
		return false
	}

	// https://en.wikipedia.org/wiki/Computer_terminal#Dumb_terminals
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	// https://no-color.org/
	if _, nocolor := os.LookupEnv("NO_COLOR"); nocolor {
		return false
	}

	// On Windows WT_SESSION is set by the modern terminal component.
	// Older terminals have poor support for UTF-8, VT escape codes, etc.
	if runtime.GOOS == "windows" && os.Getenv("WT_SESSION") == "" {
		return false
	}

	return true
}
