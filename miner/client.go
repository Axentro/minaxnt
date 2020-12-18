package miner

import (
	"minaxnt/types"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	sync.RWMutex
	Conn  *websocket.Conn
	Send  chan types.MessageResponse
	Done  chan struct{}
	block types.MinerBlock
}

func (c *Client) UpdateBlock(block types.MinerBlock) {
	c.Lock()
	defer c.Unlock()
	c.block = block
}

func (c *Client) Block() types.MinerBlock {
	c.RLock()
	defer c.RUnlock()
	return c.block
}
