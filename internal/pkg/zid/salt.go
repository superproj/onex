// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package zid

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"os"
)

func Salt() uint64 {
	// Calculate the hash value of the string using the FNV-1a hash algorithm
	h := fnv.New64a()
	h.Write(ReadMachineID())

	// Convert the hash value to a salt of type uint64
	hash := h.Sum64()
	return hash
}

func ReadMachineID() []byte {
	id := make([]byte, 3)
	hid, err := readPlatformMachineID()
	if err != nil || len(hid) == 0 {
		hid, err = os.Hostname()
	}
	if err == nil && len(hid) != 0 {
		hw := sha256.New()
		hw.Write([]byte(hid))
		copy(id, hw.Sum(nil))
	} else {
		// Fallback to rand number if machine id can't be gathered
		if _, randErr := rand.Reader.Read(id); randErr != nil {
			panic(fmt.Errorf("id: cannot get hostname nor generate a random number: %w; %w", err, randErr))
		}
	}
	return id
}

func readPlatformMachineID() (string, error) {
	b, err := ioutil.ReadFile("/etc/machine-id")
	if err != nil || len(b) == 0 {
		b, err = ioutil.ReadFile("/sys/class/dmi/id/product_uuid")
	}
	return string(b), err
}
