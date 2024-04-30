// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"fmt"
	"net"
	"strings"

	netutils "k8s.io/utils/net"
)

// Define unit constant.
const (
	_   = iota // ignore onex.iota
	KiB = 1 << (10 * iota)
	MiB
	GiB
	TiB
)

func join(prefixes ...string) string {
	joined := strings.Join(prefixes, ".")
	if joined != "" {
		joined += "."
	}

	return joined
}

// ValidateAddress takes an address as a string and validates it.
// If the input address is not in a valid :port or IP:port format, it returns an error.
// It also checks if the host part of the address is a valid IP address and if the port number is valid.
func ValidateAddress(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("%q is not in a valid format (:port or ip:port): %w", addr, err)
	}
	if host != "" && netutils.ParseIPSloppy(host) == nil {
		return fmt.Errorf("%q is not a valid IP address", host)
	}
	if _, err := netutils.ParsePort(port, true); err != nil {
		return fmt.Errorf("%q is not a valid number", port)
	}

	return nil
}

// CreateListener create net listener by given address and returns it and port.
func CreateListener(addr string) (net.Listener, int, error) {
	network := "tcp"

	ln, err := net.Listen(network, addr)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to listen on %v: %w", addr, err)
	}

	// get port
	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		_ = ln.Close()

		return nil, 0, fmt.Errorf("invalid listen address: %q", ln.Addr().String())
	}

	return ln, tcpAddr.Port, nil
}
