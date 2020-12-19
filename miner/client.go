package miner

import (
	"encoding/json"
	"fmt"
	"minaxnt/types"
	"time"

	"github.com/alitto/pond"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	Name       string
	Conn       *websocket.Conn
	Send       chan types.MessageResponse
	Done       chan struct{}
	MinerId    string
	Address    string
	Process    int
	StopMining chan int
}

func (c *Client) Start() {
	go c.send()
	go c.recv()

	// Handshake
	handshake := types.MessageResponse{
		Type:    types.TYPE_MINER_HANDSHAKE,
		Content: fmt.Sprintf("{\"version\":%d,\"address\":\"%s\",\"mid\":\"%s\"}", types.CORE_VERSION, c.Address, c.MinerId),
	}
	c.Send <- handshake
}

func (c *Client) FoundNonce(resp types.PeerResponse) {
	for {
		log.Debugf("Start mining block index: %d", resp.Block.Index)
		blockNonce, stop := Mining(resp.Block, resp.MiningDifficulty, c)
		if stop {
			return
		}
		go func() {
			log.Infof("Found new nonce(%d): %d", resp.MiningDifficulty, blockNonce)
			log.Debugf("=> Found block: %d", resp.Block.Index)

			mnc := types.MinerNonceContent{
				Nonce: types.NewMinerNonce(),
			}
			mnc.Nonce.Mid = c.MinerId
			mnc.Nonce.Value = fmt.Sprintf("%d", blockNonce)
			mnc.Nonce.Timestamp = time.Now().UTC().UnixNano() / int64(time.Millisecond)
			mnc.Nonce.Address = c.Address
			mncJSON, err := json.Marshal(mnc)
			if err != nil {
				log.Error("Can't convert miner nonce to JSON: ", err)
			}
			resultNonce := types.MessageResponse{
				Type:    types.TYPE_MINER_FOUND_NONCE,
				Content: string(mncJSON),
			}
			c.Send <- resultNonce
		}()
	}
}

func (c *Client) send() {
	for {
		select {
		case data, ok := <-c.Send:
			if !ok {
				log.Error("sendDataChan error")
				c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
			err := c.Conn.WriteJSON(&data)
			if err != nil {
				log.Error("Can't send data to the blockchain:", err)
				return
			}
		}
	}
}

func (c *Client) recv() {
	pool := pond.New(c.Process, 0, pond.MinWorkers(c.Process))
	for {
		log.Debug("Waiting for blockchain data...")

		result := types.MessageResponse{}
		err := c.Conn.ReadJSON(&result)
		if err != nil {
			log.Error("Can't retrieve handshake data: ", err)
			return
		}
		log.Debugf("Received message from blockchain: %+v", result)
		switch result.Type {
		case types.TYPE_MINER_HANDSHAKE_ACCEPTED:
			log.Debug("[MINER_HANDSHAKE_ACCEPTED]")

			resp := types.PeerResponse{}
			err = json.Unmarshal([]byte(result.Content), &resp)
			if err != nil {
				log.Error("Can't parse mining block data: ", err)
				return
			}
			log.Infof("PREPARING NEXT SLOW BLOCK: %d at approximate difficulty: %d", resp.Block.Index, resp.Block.Difficulty)

			for i := 0; i < c.Process; i++ {
				pool.Submit(func() {
					c.FoundNonce(resp)
				})
			}
		case types.TYPE_MINER_HANDSHAKE_REJECTED:
			reason := types.PeerRejectedResponse{}
			err = json.Unmarshal([]byte(result.Content), &reason)
			if err != nil {
				log.Error("Can't convert rejected message")
			}
			log.Fatal("Handshake rejected: ", reason.Reason)
		case types.TYPE_MINER_BLOCK_UPDATE:
			log.Debug("[MINER_BLOCK_UPDATE]")

			resp := types.PeerResponse{}
			err = json.Unmarshal([]byte(result.Content), &resp)
			if err != nil {
				log.Error("Can't parse mining block data: ", err)
				return
			}
			log.Infof("PREPARING NEXT SLOW BLOCK: %d at approximate difficulty: %d", resp.Block.Index, resp.Block.Difficulty)

			for i := 0; i < c.Process; i++ {
				c.StopMining <- 1
			}
			pool.StopAndWait()
			pool = pond.New(c.Process, 0, pond.MinWorkers(c.Process))

			for i := 0; i < c.Process; i++ {
				pool.Submit(func() {
					c.FoundNonce(resp)
				})
			}
		}
	}
}
