// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//nolint:errchkjson,errcheck
package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	"golang.org/x/net/websocket"

	"github.com/superproj/onex/internal/toyblc/blc"
	"github.com/superproj/onex/pkg/log"
)

type ByIndex []*blc.Block

func (b ByIndex) Len() int           { return len(b) }
func (b ByIndex) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByIndex) Less(i, j int) bool { return b[i].Index < b[j].Index }

func ConnectToPeers(ctx context.Context, bs *blc.BlockSet, ss *Sockets, peers []string) {
	for _, peer := range peers {
		if peer == "" {
			continue
		}

		ws, err := websocket.Dial(peer, "", peer)
		if err != nil {
			log.C(ctx).Errorw(err, "Dial to peer", "peer", peer)
			continue
		}

		go WSHandler(bs, ss, ws)

		log.C(ctx).Debugw("Query latest block")
		ws.Write(bs.LatestMessage())
	}
}

func WSHandler(bs *blc.BlockSet, ss *Sockets, ws *websocket.Conn) {
	var (
		resp = &blc.ResponseBlockchain{}
		peer = ws.LocalAddr().String()
	)

	ss.Add(ws)

	for {
		var msg []byte
		if err := websocket.Message.Receive(ws, &msg); err != nil {
			if errors.Is(err, io.EOF) {
				log.Warnw("P2P peer shutdown, remove it from the peers pool", "peer", peer)
				break
			}

			log.Errorw(err, "Unable to receive P2P message from", "peer", peer)
			break
		}

		log.Debugw("Received message", "peer", peer, "message", msg)
		if err := json.Unmarshal(msg, resp); err != nil {
			log.Warnw("Invalid P2P message", "err", err)
		}

		switch resp.Type {
		case blc.QueryLatestAction:
			resp.Type = blc.ResponseAction

			message := bs.LatestMessage()
			log.Debugw("Responding with the latest message", "message", message)
			ws.Write(message)

		case blc.QueryAllAction:
			resp.Type = blc.ResponseAction
			resp.Data, _ = bs.MarshalJSON()
			data, _ := json.Marshal(resp)
			log.Debugw("Responding with the chain message", "message", data)
			ws.Write(data)

		case blc.ResponseAction:
			ResponseBlockchain(bs, ss, resp.Data)
		}
	}
}

func ResponseBlockchain(bs *blc.BlockSet, ss *Sockets, msg []byte) {
	receivedBlocks := []*blc.Block{}

	if err := json.Unmarshal(msg, &receivedBlocks); err != nil {
		log.Warnw("Invalid blockchain", "err", err)
	}

	sort.Sort(ByIndex(receivedBlocks))

	latestBlockReceived := receivedBlocks[len(receivedBlocks)-1]
	latestBlockHeld := bs.Latest()
	if latestBlockReceived.Index <= latestBlockHeld.Index {
		log.Infow("Received blockchain is not longer than the current blockchain. No action needed")
		return
	}

	log.Warnf("Blockchain may be behind. We have: %d Peer has: %d", latestBlockHeld.Index, latestBlockReceived.Index)
	if latestBlockHeld.Hash == latestBlockReceived.PreviousHash {
		log.Infof("We can append the received block to our chain")
		bs.Add(latestBlockReceived)
	} else if len(receivedBlocks) == 1 {
		log.Infow("We need to query the chain from our peer")
		ss.Broadcast(queryAllMsg())
	} else {
		log.Infow("Received blockchain is longer than the current blockchain")
		replaceBlocks(receivedBlocks, bs, ss)
	}
}

func queryAllMsg() []byte {
	return []byte(fmt.Sprintf("{\"type\": %d}", blc.QueryAllAction))
}

func replaceBlocks(src []*blc.Block, dst *blc.BlockSet, ss *Sockets) {
	if !blc.IsValidChain(src) || len(src) <= dst.Len() {
		log.Errorf("Received blockchain is invalid")
		return
	}

	log.Debugw("Received blockchain is valid. Replacing the current blockchain with the received blockchain. Peer disconnected")
	dst.SetBlocks(src)
	ss.Broadcast(dst.LatestMessage())
}
