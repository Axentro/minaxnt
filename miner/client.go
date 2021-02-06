package miner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"minaxnt/types"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/alitto/pond"
	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/klauspost/cpuid/v2"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	sync.Mutex
	ClientName     string
	CPUModel       string
	CPUFeatures    string
	CPUCores       string
	CPUCaches      string
	NodeURL        string
	conn           *websocket.Conn
	sendChan       chan *types.MessageResponse
	MinerID        string
	Address        string
	Process        int
	StopMiningChan chan bool
	Stats          *Stats
	pool           *pond.WorkerPool
	handshake      *types.MessageResponse
}

func NewClient(clientName string, nodeURL string, walletAddr string, numProcess int) *Client {
	minerID := strings.Replace(uuid.New().String(), "-", "", -1)
	return &Client{
		ClientName:     clientName,
		CPUModel:       cpuModel(),
		CPUFeatures:    cpuFeatures(),
		CPUCores:       fmt.Sprintf("Physical => %d, Logical => %d, Threads/core => %d", cpuid.CPU.PhysicalCores, cpuid.CPU.LogicalCores, cpuid.CPU.ThreadsPerCore),
		CPUCaches:      fmt.Sprintf("L2 => %s, L3 => %s", humanize.Bytes(uint64(cpuid.CPU.Cache.L2)), humanize.Bytes(uint64(cpuid.CPU.Cache.L3))),
		NodeURL:        nodeURL,
		conn:           buildConn(nodeURL, false),
		sendChan:       make(chan *types.MessageResponse),
		MinerID:        minerID,
		Address:        walletAddr,
		Process:        numProcess,
		StopMiningChan: make(chan bool, numProcess),
		Stats:          NewStats(),
		pool:           pond.New(numProcess, 0, pond.MinWorkers(numProcess)),
		handshake: &types.MessageResponse{
			Type:    types.TypeMinerHandshake,
			Content: fmt.Sprintf("{\"version\":\"%s\",\"address\":\"%s\",\"mid\":\"%s\"}", types.CoreVersion, walletAddr, minerID),
		},
	}
}

func cpuModel() string {
	cpuModel := cpuid.CPU.BrandName

	if len(cpuModel) == 0 {
		cpuModel = "[-]"
	}

	return cpuModel
}

func cpuFeatures() string {
	var cpuFeaturesBuffer bytes.Buffer

	if cpuid.CPU.Has(cpuid.SSE) {
		cpuFeaturesBuffer.WriteString("SSE:[✔]")
	}
	if cpuid.CPU.Has(cpuid.SSE2) {
		if cpuFeaturesBuffer.Len() != 0 {
			cpuFeaturesBuffer.WriteString(", ")
		}
		cpuFeaturesBuffer.WriteString("SSE2:[✔]")
	}
	if cpuid.CPU.Has(cpuid.SSE4) {
		if cpuFeaturesBuffer.Len() != 0 {
			cpuFeaturesBuffer.WriteString(", ")
		}
		cpuFeaturesBuffer.WriteString("SSE4:[✔]")
	}
	if cpuid.CPU.Has(cpuid.AVX) {
		if cpuFeaturesBuffer.Len() != 0 {
			cpuFeaturesBuffer.WriteString(", ")
		}
		cpuFeaturesBuffer.WriteString("AVX:[✔]")
	}
	if cpuid.CPU.Has(cpuid.AVX2) {
		if cpuFeaturesBuffer.Len() != 0 {
			cpuFeaturesBuffer.WriteString(", ")
		}
		cpuFeaturesBuffer.WriteString("AVX2:[✔]")
	}

	if cpuFeaturesBuffer.Len() == 0 {
		cpuFeaturesBuffer.WriteString("[-]")
	}
	return cpuFeaturesBuffer.String()
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

func (c *Client) stopMining() {
	for i := 0; i < c.Process; i++ {
		c.StopMiningChan <- true
	}
	c.pool.StopAndWait()
}

func (c *Client) foundNonce(resp types.PeerResponse, workerID int) {
	log.Debugf("[#%d] Start mining block index: %d", workerID, resp.Block.Index)
	for {
		mr, stop := Mining(resp.Block, resp.MiningDifficulty, c)
		if stop {
			log.Debugf("[#%d] Stop mining block index: %d", workerID, resp.Block.Index)
			return
		}

		log.Infof("[#%d] Found new nonce(diff. %d, required %d): %s", workerID, mr.Difficulty, resp.MiningDifficulty, mr.Nonce)
		log.Debugf("[#%d] => Nonce for block: %d", workerID, resp.Block.Index)

		mnc := types.MinerNonceContent{
			Nonce: types.NewMinerNonce(),
		}
		mnc.Nonce.Mid = c.MinerID
		mnc.Nonce.Value = mr.Nonce
		mnc.Nonce.Timestamp = mr.Timestamp
		mnc.Nonce.Address = c.Address
		mnc.Nonce.Difficulty = resp.MiningDifficulty

		mncJSON, err := json.Marshal(mnc)
		if err != nil {
			log.Errorf("[#%d] Can't convert miner nonce to JSON: %s", workerID, err)
		}

		resultNonce := types.MessageResponse{
			Type:    types.TypeMinerFoundNonce,
			Content: string(mncJSON),
		}

		c.sendChan <- &resultNonce
		log.Debugf("Miner nonce content sent to node: %v", resultNonce)
	}
}

func (c *Client) resetConnOrFail() {
	c.stopMining()

	_ = c.conn.Close()
	c.conn = buildConn(c.NodeURL, true)
	log.Warn("Node connection established from error")

	c.sendHandshake()
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
		log.Debug("Waiting for node data...")

		result := types.MessageResponse{}
		err := c.conn.ReadJSON(&result)
		if err != nil {
			c.handleError(err)
		}
		log.Debugf("Received message from blockchain: %+v", result)

		switch result.Type {
		case types.TypeMinerBlockInvalid:
			c.stopMining()
			
			log.Debug("[MINER_BLOCK_INVALID]")

			resp := types.PeerResponseWithReason{}
			err = json.Unmarshal([]byte(result.Content), &resp)
			if err != nil {
				log.Error("Can't parse mining block data: ", err)
				return
			}
			log.Warnf("[MINING BLOCK INVALID]: %s", resp.Reason)
			log.Warnf("[MINING BLOCK UPDATE (last was invalid)]: block index %d at approximate difficulty: %d", resp.Block.Index, resp.Block.Difficulty)
			for i := 0; i < c.Process; i++ {
				func(workerID int) {
					c.pool.Submit(func() {
						c.foundNonce(resp.ToPeerResponse(), workerID)
					})
				}(i)
			}
		case types.TypeMinerBlockDifficultyAdjust:
			c.stopMining()

			log.Debug("[MINER_BLOCK_DIFFICULTY_ADJUST]")

			resp := types.PeerResponseWithReason{}
			err = json.Unmarshal([]byte(result.Content), &resp)
			if err != nil {
				log.Error("Can't parse mining block data: ", err)
				return
			}
			log.Infof("[MINING DIFFICULTY ADJUST]: %s", resp.Reason)
			log.Infof("=> [BLOCK INFO]: block index %d at approximate difficulty: %d", resp.Block.Index, resp.Block.Difficulty)
			for i := 0; i < c.Process; i++ {
				func(workerID int) {
					c.pool.Submit(func() {
						c.foundNonce(resp.ToPeerResponse(), workerID)
					})
				}(i)
			}
		case types.TypeMinerBlockUpdate:
			c.stopMining()

			log.Debug("[MINER_BLOCK_UPDATE]")

			resp := types.PeerResponse{}
			err = json.Unmarshal([]byte(result.Content), &resp)
			if err != nil {
				log.Error("Can't parse mining block data: ", err)
				return
			}
			log.Infof("[NEW BLOCK]: block index %d at approximate difficulty: %d", resp.Block.Index, resp.Block.Difficulty)
			for i := 0; i < c.Process; i++ {
				func(workerID int) {
					c.pool.Submit(func() {
						c.foundNonce(resp, workerID)
					})
				}(i)
			}
		case types.TypeMinerHandshakeAccepted:
			log.Debug("[MINER_HANDSHAKE_ACCEPTED]")

			resp := types.PeerResponse{}
			err = json.Unmarshal([]byte(result.Content), &resp)
			if err != nil {
				log.Error("Can't parse mining block data: ", err)
				return
			}
			log.Infof("[START MINING]: block index %d at approximate difficulty: %d", resp.Block.Index, resp.Block.Difficulty)

			for i := 0; i < c.Process; i++ {
				func(workerID int) {
					c.pool.Submit(func() {
						c.foundNonce(resp, workerID)
					})
				}(i)
			}
		case types.TypeMinerHandshakeRejected:
			reason := types.PeerRejectedResponse{}
			err = json.Unmarshal([]byte(result.Content), &reason)
			if err != nil {
				log.Error("Can't convert rejected message")
			}
			log.Fatal("Handshake rejected: ", reason.Reason)
		default:
			log.Warnf("Unknonw response type %d: %v", result.Type, result)
		}
	}
}
