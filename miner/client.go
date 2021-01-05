package miner

import (
	"encoding/json"
	"fmt"
	"minaxnt/types"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alitto/pond"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	sync.Mutex
	ClientName string
	NodeURL    string
	Conn       *websocket.Conn
	Send       chan *types.MessageResponse
	MinerID    string
	Address    string
	Process    int
	StopMining chan struct{}
	Stats      *Stats
}

func NewClient(clientName string, nodeURL string, walletAddr string, numProcess int) *Client {
	return &Client{
		ClientName: clientName,
		NodeURL:    nodeURL,
		Conn:       buildConn(nodeURL),
		Send:       make(chan *types.MessageResponse),
		MinerID:    strings.Replace(uuid.New().String(), "-", "", -1),
		Address:    walletAddr,
		Process:    numProcess,
		StopMining: make(chan struct{}, numProcess),
		Stats:      NewStats(),
	}
}

func buildConn(nodeURL string) *websocket.Conn {
	nu, err := url.Parse(nodeURL)
	if err != nil {
		log.Fatal("Can't parse the node URL: ", err)
	}
	scheme := "ws"
	if nu.Scheme == "https" {
		scheme = "wss"
	}
	u := url.URL{Scheme: scheme, Host: nu.Host, Path: "/peer"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	return conn
}

func (c *Client) Start() {
	go c.send()
	go c.recv()
	go c.Stats.Start()

	// Handshake
	handshake := types.MessageResponse{
		Type:    types.TYPE_MINER_HANDSHAKE,
		Content: fmt.Sprintf("{\"version\":%d,\"address\":\"%s\",\"mid\":\"%s\"}", types.CORE_VERSION, c.Address, c.MinerID),
	}
	c.Send <- &handshake
}

func (c *Client) FoundNonce(resp types.PeerResponse, Id int) {
	for {
		log.Debugf("[#%d] Start mining block index: %d", Id, resp.Block.Index)
		blockNonce, computedDifficulty, stop := Mining(resp.Block, resp.MiningDifficulty, c)
		if stop {
			log.Debugf("[#%d] Stop mining block index: %d", Id, resp.Block.Index)
			return
		}
		go func() {
			log.Infof("[#%d] Found new nonce(diff. %d, required %d): %d", Id, computedDifficulty, resp.MiningDifficulty, blockNonce)
			log.Debugf("[#%d] => Nonce for block: %d", Id, resp.Block.Index)

			mnc := types.MinerNonceContent{
				Nonce: types.NewMinerNonce(),
			}
			mnc.Nonce.Mid = c.MinerID
			mnc.Nonce.Value = strconv.Itoa(int(blockNonce))
			mnc.Nonce.Timestamp = time.Now().UTC().UnixNano() / int64(time.Millisecond)
			mnc.Nonce.Address = c.Address
			mncJSON, err := json.Marshal(mnc)
			if err != nil {
				log.Errorf("[#%d] Can't convert miner nonce to JSON: %s", Id, err)
			}
			resultNonce := types.MessageResponse{
				Type:    types.TYPE_MINER_FOUND_NONCE,
				Content: string(mncJSON),
			}
			c.Send <- &resultNonce
		}()
	}
}

func (c *Client) resetConn() {
	c.Lock()
	defer c.Unlock()
	_ = c.Conn.Close()
	c.Conn = buildConn(c.NodeURL)
}

func (c *Client) handleError(err error) {
	switch {
	case websocket.IsCloseError(err, websocket.CloseNormalClosure):
		log.Fatalf("Normal node closure: %v", err)
	case websocket.IsCloseError(err, websocket.CloseAbnormalClosure):
		log.Fatalf("Node connection closed abnormally: %v", err)
	case websocket.IsCloseError(err, websocket.CloseTryAgainLater):
		log.Warn("Node is online but not ready")
		log.Debug("=> Waiting for node for 60 seconds")
		time.Sleep(60 * time.Second)
		c.resetConn()
		log.Debug("=> Node connection reseted")
	default:
		log.Fatalf("Connection is not available: %v", err)
	}
}

func (c *Client) send() {
	for {
		data, ok := <-c.Send
		if !ok {
			_ = c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
		err := c.Conn.WriteJSON(data)
		if err != nil {
			c.handleError(err)
		}
	}

}

func (c *Client) recv() {
	pool := pond.New(c.Process, 0, pond.MinWorkers(c.Process))
	for {
		log.Debug("Waiting for node data...")

		result := types.MessageResponse{}
		err := c.Conn.ReadJSON(&result)
		if err != nil {
			c.handleError(err)
		}
		log.Debugf("Received message from blockchain: %+v", result)
		switch result.Type {
		case types.TYPE_MINER_BLOCK_UPDATE:
			log.Debug("[MINER_BLOCK_UPDATE]")
			for i := 0; i < c.Process; i++ {
				c.StopMining <- struct{}{}
			}
			pool.StopAndWait()

			resp := types.PeerResponse{}
			err = json.Unmarshal([]byte(result.Content), &resp)
			if err != nil {
				log.Error("Can't parse mining block data: ", err)
				return
			}
			log.Infof("[BLOCK-UPDATE] PREPARING NEXT SLOW BLOCK: %d at approximate difficulty: %d", resp.Block.Index, resp.Block.Difficulty)
			for i := 0; i < c.Process; i++ {
				func(Id int) {
					pool.Submit(func() {
						c.FoundNonce(resp, Id)
					})
				}(i)
			}
		case types.TYPE_MINER_HANDSHAKE_ACCEPTED:
			log.Debug("[MINER_HANDSHAKE_ACCEPTED]")

			resp := types.PeerResponse{}
			err = json.Unmarshal([]byte(result.Content), &resp)
			if err != nil {
				log.Error("Can't parse mining block data: ", err)
				return
			}
			log.Infof("[START] PREPARING NEXT SLOW BLOCK: %d at approximate difficulty: %d", resp.Block.Index, resp.Block.Difficulty)

			for i := 0; i < c.Process; i++ {
				func(Id int) {
					pool.Submit(func() {
						c.FoundNonce(resp, Id)
					})
				}(i)
			}
		case types.TYPE_MINER_HANDSHAKE_REJECTED:
			reason := types.PeerRejectedResponse{}
			err = json.Unmarshal([]byte(result.Content), &reason)
			if err != nil {
				log.Error("Can't convert rejected message")
			}
			log.Fatal("Handshake rejected: ", reason.Reason)
		}
	}
}
