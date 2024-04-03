// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package blc

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/wire"

	"github.com/superproj/onex/internal/toyblc/defaults"
	"github.com/superproj/onex/pkg/log"
)

//nolint:errchkjson
type Action int

const (
	QueryLatestAction Action = iota
	QueryAllAction
	ResponseAction
)

var ProviderSet = wire.NewSet(NewBlockSet)

var genesis = &Block{
	Index:        0,
	PreviousHash: "0",
	Timestamp:    1465154705,
	Data:         "genesis block",
	Hash:         "816534932c2b7154836da6afc367695e6337db8a921823784c14378abed4f7d7",
	Address:      defaults.GenesisAddress,
}

type Block struct {
	Index        int64  `json:"index"`
	PreviousHash string `json:"previousHash"`
	Timestamp    int64  `json:"timestamp"`
	Data         string `json:"data"`
	Hash         string `json:"hash"`
	Address      string `json:"address"`
}

func (b *Block) String() string {
	return fmt.Sprintf("index: %d,previousHash:%s,timestamp:%d,data:%s,hash:%s", b.Index, b.PreviousHash, b.Timestamp, b.Data, b.Hash)
}

func (b *Block) CalHash() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d%s%d%s", b.Index, b.PreviousHash, b.Timestamp, b.Data))))
}

type ResponseBlockchain struct {
	Type Action `json:"type"`
	Data []byte `json:"data"`
}

type BlockSet struct {
	address string
	data    []*Block
}

func NewBlockSet(address string) *BlockSet {
	return &BlockSet{
		address: address,
		data:    []*Block{genesis},
	}
}

func (bs *BlockSet) List() []*Block {
	return bs.data
}

func (bs *BlockSet) Add(b *Block) {
	if !isValidNewBlock(b, bs.Latest()) {
		return
	}

	bs.data = append(bs.data, b)
}

func (bs *BlockSet) Latest() *Block {
	return bs.data[len(bs.data)-1]
}

func (bs *BlockSet) Len() int {
	return len(bs.data)
}

func (bs *BlockSet) SetBlocks(blocks []*Block) {
	bs.data = blocks
}

func (bs *BlockSet) LatestMessage() []byte {
	data, _ := json.Marshal([]*Block{bs.Latest()})
	resp := &ResponseBlockchain{
		Type: ResponseAction,
		Data: data,
	}

	data, _ = json.Marshal(resp)
	return data
}

func (bs *BlockSet) NextBlock(data string) *Block {
	pre := bs.Latest()

	nb := &Block{
		Data:         data,
		PreviousHash: pre.Hash,
		Index:        pre.Index + 1,
		Timestamp:    time.Now().Unix(),
		Address:      bs.address,
	}

	nb.Hash = nb.CalHash()

	return nb
}

func (bs *BlockSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(bs)
}

func isValidNewBlock(nb, pb *Block) bool {
	if nb.Hash == nb.CalHash() && pb.Index+1 == nb.Index && pb.Hash == nb.PreviousHash {
		return true
	}

	return false
}

func IsValidChain(blocks []*Block) bool {
	if len(blocks) == 0 {
		return false
	}

	if blocks[0].String() != genesis.String() {
		log.Warnw("No matching GenesisBlock", "block", blocks[0].String())
		return false
	}

	temp := []*Block{blocks[0]}
	for i := 1; i < len(blocks); i++ {
		if isValidNewBlock(blocks[i], temp[i-1]) {
			return false
		}

		temp = append(temp, blocks[i])
	}

	return true
}

type ByIndex []*Block

func (b ByIndex) Len() int           { return len(b) }
func (b ByIndex) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByIndex) Less(i, j int) bool { return b[i].Index < b[j].Index }
