package main

import (
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

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	address = flag.StringP("address", "a", "", "Axentro address to receive rewards")
	node    = flag.StringP("node", "n", "http://mainnet.axentro.io", "Node URL to mine against")
	process = flag.IntP("process", "p", 1, "Number of core(s) to use")
	Version = "v0.0.0"
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

func main() {
	if *address == "" {
		flag.Usage()
		log.Fatal("Wallet address is missing !")
	}
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

	client := &miner.Client{
		Conn:       c,
		Send:       make(chan types.MessageResponse),
		Done:       make(chan struct{}),
		MinerId:    strings.Replace(uuid.New().String(), "-", "", -1),
		Address:    *address,
		Process:    *process,
		StopMining: make(chan int, *process),
	}
	defer close(client.Done)

	util.Welcome(*node, *address, client.MinerId, *process, Version)
	client.Start()

	select {
	case <-client.Done:
		return
	case <-interrupt:
		log.Warn("MinAXNT interrupt!!!")
		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Error("Error when sending close message to the blockchain:", err)
		}
		select {
		case <-time.After(time.Second):
		}
		return
	}
}
