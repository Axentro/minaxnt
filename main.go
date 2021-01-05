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

	"github.com/mattn/go-colorable"
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
	log.SetOutput(colorable.NewColorableStdout())
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
		flag.Usage()
		log.Fatal("Wallet address is missing or is not valid!")
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	client := miner.NewClient(fmt.Sprintf("%s - %s", MinerName, Version), *node, *address, *process)
	util.Welcome(client)
	client.Start()

	<-interrupt
	log.Warnf("%s is stopped", client.ClientName)
	time.Sleep(1 * time.Second) // A chance for remote close signal to be send
}
