package main

import (
	"fmt"
	"math/rand"
	"minaxnt/miner"
	"minaxnt/util"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

var (
	address   = flag.StringP("address", "a", "", "Axentro address to receive rewards")
	node      = flag.StringP("node", "n", "http://mainnet.axentro.io", "Node URL to mine against")
	process   = flag.IntP("process", "p", 1, "Number of core(s) to use")
	debug     = flag.Bool("debug", false, "Set log level to debug")
	MinerName = "MinAXNT"
	Version   = "v0.0.0"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().Unix())
	logrus.SetOutput(colorable.NewColorableStdout())
	log.SetFormatter(&log.TextFormatter{
		DisableColors:          false,
		DisableLevelTruncation: true,
		ForceColors:            true,
		FullTimestamp:          true,
	})
	flag.Parse()
}

func main() {
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	if !util.IsValidAddress(*address) {
		log.Fatal("Wallet address is missing or is not valid!")
		flag.Usage()
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	client := miner.NewClient(fmt.Sprintf("%s - %s", MinerName, Version), *node, *address, *process)
	util.Welcome(client)
	client.Start()

	select {
	case <-client.Done:
		_ = client.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		client.Conn.Close()
		return
	case <-interrupt:
		log.Warnf("%s interrupt!!!", client.ClientName)
		err := client.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Errorf("Error when sending close message to the blockchain: %s", err)
		}
		select {
		case <-time.After(1 * time.Second):
		}
		return
	}
}
