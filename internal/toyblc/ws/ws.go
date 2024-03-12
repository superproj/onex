// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//nolint:errchkjson
package ws

import (
	"encoding/json"

	"github.com/google/wire"
	"golang.org/x/net/websocket"

	"github.com/superproj/onex/pkg/log"
)

type Sockets []*websocket.Conn

var ProviderSet = wire.NewSet(NewSockets)

func NewSockets() *Sockets {
	return new(Sockets)
}

func (ss *Sockets) String() string {
	data, _ := json.Marshal(ss)
	return string(data)
}

func (ss *Sockets) Broadcast(msg []byte) {
	for n, socket := range *ss {
		if _, err := socket.Write(msg); err != nil {
			log.Warnw("Peer disconnected", "peer", socket.RemoteAddr().String())
			*ss = append((*ss)[0:n], (*ss)[n+1:]...)
		}
	}
}

func (ss *Sockets) List() []*websocket.Conn {
	return []*websocket.Conn(*ss)
}

func (ss *Sockets) Add(ws *websocket.Conn) {
	*ss = append(*ss, ws)
}
