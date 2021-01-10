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
	"github.com/klauspost/cpuid/v2"
	log "github.com/sirupsen/logrus"
	"github.com/tevino/abool"
)

type Client struct {
	sync.Mutex
	ClientName  string
	CPUModel    string
	CPUFeatures string
	CPUCores    string
	NodeURL     string
	conn        *websocket.Conn
	sendChan    chan *types.MessageResponse
	MinerID     string
	Address     string
	Process     int
	StopClient  *abool.AtomicBool
	Stats       *Stats
	pool        *pond.WorkerPool
	handshake   *types.MessageResponse
}

func NewClient(clientName string, nodeURL string, walletAddr string, numProcess int) *Client {
	minerID := strings.Replace(uuid.New().String(), "-", "", -1)
	cpuFeatures := "[-]"
	if cpuid.CPU.Supports(cpuid.SSE, cpuid.SSE2, cpuid.SSE4, cpuid.AVX, cpuid.AVX2) {
		cpuFeatures = "SSE, SSE2, SSE4: [✔] - AVX, AVX2: [✔]"
	}
	return &Client{
		ClientName:  clientName,
		CPUModel:    cpuid.CPU.BrandName,
		CPUFeatures: cpuFeatures,
		CPUCores:    fmt.Sprintf("Physical => %d, Logical => %d, Threads/core => %d", cpuid.CPU.PhysicalCores, cpuid.CPU.LogicalCores, cpuid.CPU.ThreadsPerCore),
		NodeURL:     nodeURL,
		conn:        buildConn(nodeURL, false),
		sendChan:    make(chan *types.MessageResponse),
		MinerID:     minerID,
		Address:     walletAddr,
		Process:     numProcess,
		StopClient:  abool.New(),
		Stats:       NewStats(),
		pool:        pond.New(numProcess, 0, pond.MinWorkers(numProcess)),
		handshake: &types.MessageResponse{
			Type:    types.TYPE_MINER_HANDSHAKE,
			Content: fmt.Sprintf("{\"version\":%d,\"address\":\"%s\",\"mid\":\"%s\"}", types.CORE_VERSION, walletAddr, minerID),
		},
	}
}

func buildConn(nodeURL string, retry bool) *websocket.Conn {
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
		if retry {
			retryTimes := 180
			retrySleep := 10 * time.Second
			for {
				retryTimes--
				time.Sleep(retrySleep)
				conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
				if err != nil {
					if retryTimes == 0 {
						log.Fatal("Tried to reconnect without success")
					}
					log.Warn("Failed to reconnect to the node, trying again...")
					continue
				}
				log.Warn("Connection to the node is establiched again")
				break
			}
		} else {
			log.Fatalf("Connection to node failed: %v", err)
		}
	}
	return conn
}

func (c *Client) sendHandshake() {
	c.sendChan <- c.handshake
}

func (c *Client) Start() {
	go c.Stats.Start()
	go c.send()
	go c.recv()

	c.sendHandshake()
}

func (c *Client) startMining() {
	c.StopClient.UnSet()
}

func (c *Client) stopMining() {
	c.StopClient.Set()
	c.pool.StopAndWait()
}

func (c *Client) foundNonce(resp types.PeerResponse, workerID int) {
	for {
		log.Debugf("[#%d] Start mining block index: %d", workerID, resp.Block.Index)
		blockNonce, computedDifficulty, stop := Mining(resp.Block, resp.MiningDifficulty, c)
		if stop {
			log.Debugf("[#%d] Stop mining block index: %d", workerID, resp.Block.Index)
			return
		}
		go func() {
			log.Infof("[#%d] Found new nonce(diff. %d, required %d): %d", workerID, computedDifficulty, resp.MiningDifficulty, blockNonce)
			log.Debugf("[#%d] => Nonce for block: %d", workerID, resp.Block.Index)

			mnc := types.MinerNonceContent{
				Nonce: types.NewMinerNonce(),
			}
			mnc.Nonce.Mid = c.MinerID
			mnc.Nonce.Value = strconv.Itoa(int(blockNonce))
			mnc.Nonce.Timestamp = time.Now().UTC().UnixNano() / int64(time.Millisecond)
			mnc.Nonce.Address = c.Address
			mncJSON, err := json.Marshal(mnc)
			if err != nil {
				log.Errorf("[#%d] Can't convert miner nonce to JSON: %s", workerID, err)
			}
			resultNonce := types.MessageResponse{
				Type:    types.TYPE_MINER_FOUND_NONCE,
				Content: string(mncJSON),
			}
			c.sendChan <- &resultNonce
		}()
	}
}

func (c *Client) resetConnOrFail() {
	c.stopMining()

	_ = c.conn.Close()
	c.conn = buildConn(c.NodeURL, true)
	log.Warn("Node connection established from error")

	c.sendHandshake()

	c.startMining()
}

func (c *Client) handleError(err error) {
	c.Lock()
	defer c.Unlock()

	switch {
	case websocket.IsCloseError(err, websocket.CloseNormalClosure):
		log.Errorf("Normal node closure: %v", err)
	case websocket.IsCloseError(err, websocket.CloseAbnormalClosure):
		log.Errorf("Node connection closed abnormally: %v", err)
	case websocket.IsCloseError(err, websocket.CloseTryAgainLater):
		log.Warn("Node is online but not ready")
		log.Debug("=> Waiting for node for 60 seconds")
		time.Sleep(60 * time.Second)
	default:
		log.Errorf("Node connection lost: %v", err)
	}
	c.resetConnOrFail()
	log.Debug("=> Node connection is back again")
}

func (c *Client) send() {
	for {
		if c.StopClient.IsSet() {
			time.Sleep(1 * time.Second)
			continue
		}

		data, ok := <-c.sendChan
		if !ok {
			_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
		err := c.conn.WriteJSON(data)
		if err != nil {
			c.handleError(err)
		}
	}
}

func (c *Client) recv() {
	for {
		if c.StopClient.IsSet() {
			time.Sleep(1 * time.Second)
			continue
		}

		log.Debug("Waiting for node data...")

		result := types.MessageResponse{}
		err := c.conn.ReadJSON(&result)
		if err != nil {
			c.handleError(err)
		}
		log.Debugf("Received message from blockchain: %+v", result)

		switch result.Type {
		case types.TYPE_MINER_BLOCK_UPDATE:
			log.Debug("[MINER_BLOCK_UPDATE]")

			c.stopMining()
			c.startMining()

			resp := types.PeerResponse{}
			err = json.Unmarshal([]byte(result.Content), &resp)
			if err != nil {
				log.Error("Can't parse mining block data: ", err)
				return
			}
			log.Infof("[BLOCK-UPDATE] PREPARING NEXT SLOW BLOCK: %d at approximate difficulty: %d", resp.Block.Index, resp.Block.Difficulty)
			for i := 0; i < c.Process; i++ {
				func(workerID int) {
					c.pool.Submit(func() {
						c.foundNonce(resp, workerID)
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
				func(workerID int) {
					c.pool.Submit(func() {
						c.foundNonce(resp, workerID)
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
