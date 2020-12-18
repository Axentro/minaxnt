package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"minaxnt/miner"
	"minaxnt/types"
	"minaxnt/util"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"

	"github.com/alitto/pond"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	address = flag.StringP("address", "a", "TTAyNmVjNGU4MTBlYzc1ZWVmNDMyYzc4MjI5NGRmZWNhYzUwMzJjN2UyYzlhNzE3", "Axentro address to receive rewards")
	node    = flag.StringP("node", "n", "http://mainnet.axentro.io", "Node URL to mine against")
	process = flag.IntP("process", "p", 1, "Number of core(s) to use")
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().Unix())
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetOutput(colorable.NewColorableStdout())
	// log.SetLevel(log.DebugLevel)
	flag.Parse()
}

type Client struct {
	conn *websocket.Conn
	send chan types.MessageResponse
	done chan struct{}
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	nodeURL, err := url.Parse(*node)
	if err != nil {
		log.Fatal("Can't parse the node URL: ", err)
	}
	scheme := "ws"
	if nodeURL.Scheme == "https" {
		scheme = "wss"
	}
	u := url.URL{Scheme: scheme, Host: nodeURL.Host, Path: "/peer"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	// Miner UUID
	minerId := strings.Replace(uuid.New().String(), "-", "", -1)
	util.Welcome(*node, *address, minerId, *process)

	client := &Client{
		conn: c,
		send: make(chan types.MessageResponse),
		done: make(chan struct{}),
	}
	defer close(client.done)

	go send(client)
	go recv(client, minerId)

	// Handshake
	handshake := types.MessageResponse{
		Type:    types.TYPE_MINER_HANDSHAKE,
		Content: fmt.Sprintf("{\"version\":%d,\"address\":\"%s\",\"mid\":\"%s\"}", types.CORE_VERSION, *address, minerId),
	}
	client.send <- handshake

	select {
	case <-client.done:
		return
	case <-interrupt:
		log.Warn("MinAXNT interrupt!!!")
		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Error("write close:", err)
		}
		select {
		case <-time.After(time.Second):
		}
		return
	}
}

func send(c *Client) {
	for {
		select {
		case data, ok := <-c.send:
			if !ok {
				log.Error("sendDataChan error")
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := c.conn.WriteJSON(&data)
			if err != nil {
				log.Error("Can't send data to the blockchain:", err)
				return
			}
		default:
		}
	}
}

func recv(c *Client, minerId string) {
	pool := pond.New(64, 128, pond.MinWorkers(*process))
	for {
		log.Debug("Waiting for blockchain data...")

		result := types.MessageResponse{}
		err := c.conn.ReadJSON(&result)
		if err != nil {
			log.Error("Can't retrieve handshake data: ", err)
			close(c.done)
			return
		}
		log.Debugf("Received message from blockchain: %+v", result)
		switch result.Type {
		case types.TYPE_MINER_HANDSHAKE_ACCEPTED:
			pool.StopAndWait()
			log.Debug("[MINER_HANDSHAKE_ACCEPTED]")
			resp := types.PeerResponse{}
			err = json.Unmarshal([]byte(result.Content), &resp)
			if err != nil {
				log.Error("Can't parse mining block data: ", err)
				return
			}
			log.Infof("PREPARING NEXT SLOW BLOCK: %d at approximate difficulty: %d", resp.Block.Index, resp.Block.Difficulty)
			for i := 0; i < *process; i++ {
				pool.Submit(func() {
					for {
						blockNonce := miner.Mining(resp.Block, resp.MiningDifficulty)
						log.Infof("Found new nonce(%d): %s", resp.MiningDifficulty, blockNonce.Nonce)
						log.Debugf("=> Found block: %+v", blockNonce)

						mnc := types.MinerNonceContent{
							Nonce: types.NewMinerNonce(),
						}
						mnc.Nonce.Mid = minerId
						mnc.Nonce.Value = blockNonce.Nonce
						mnc.Nonce.Timestamp = time.Now().UTC().UnixNano() / int64(time.Millisecond)
						mnc.Nonce.Address = *address
						mncJSON, err := json.Marshal(mnc)
						if err != nil {
							log.Error("Can't convert miner nonce to JSON: ", err)
						}
						resultNonce := types.MessageResponse{
							Type:    types.TYPE_MINER_FOUND_NONCE,
							Content: string(mncJSON),
						}
						c.send <- resultNonce
					}
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
			pool.StopAndWait()
			log.Debug("[MINER_BLOCK_UPDATE]")
			resp := types.PeerResponse{}
			err = json.Unmarshal([]byte(result.Content), &resp)
			if err != nil {
				log.Error("Can't parse mining block data: ", err)
				return
			}
			log.Infof("PREPARING NEXT SLOW BLOCK: %d at approximate difficulty: %d", resp.Block.Index, resp.Block.Difficulty)
			for i := 0; i < *process; i++ {
				pool.Submit(func() {
					for {
						blockNonce := miner.Mining(resp.Block, resp.MiningDifficulty)
						log.Infof("Found new nonce(%d): %s", resp.MiningDifficulty, blockNonce.Nonce)
						log.Debugf("=> Found block: %+v", blockNonce)

						mnc := types.MinerNonceContent{
							Nonce: types.NewMinerNonce(),
						}
						mnc.Nonce.Mid = minerId
						mnc.Nonce.Value = blockNonce.Nonce
						mnc.Nonce.Timestamp = time.Now().UTC().UnixNano() / int64(time.Millisecond)
						mnc.Nonce.Address = *address
						mncJSON, err := json.Marshal(mnc)
						if err != nil {
							log.Error("Can't convert miner nonce to JSON: ", err)
						}
						resultNonce := types.MessageResponse{
							Type:    types.TYPE_MINER_FOUND_NONCE,
							Content: string(mncJSON),
						}
						c.send <- resultNonce
					}
				})
			}
		}
	}
}
