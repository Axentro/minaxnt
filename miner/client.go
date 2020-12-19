package miner

import (
	"encoding/json"
	"fmt"
	"minaxnt/types"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	sync.RWMutex
	Conn       *websocket.Conn
	Send       chan types.MessageResponse
	Done       chan struct{}
	MinerId    string
	Address    string
	Process    int
	StopMining chan int
}

func (c *Client) Mine(resp types.PeerResponse) {
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
