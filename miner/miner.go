package miner

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"minaxnt/types"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/argon2"
)

const (
	Argon2Iterations  = 1
	Argon2Memory      = 64 * 1024
	Argon2Parallelism = 1
	Argon2KeyLength   = 512
)

func validate(blockHash string, blockNonce uint64, difficulty int32) int32 {
	nonce := strconv.FormatUint(blockNonce, 16)
	if len([]rune(nonce))%2 != 0 {
		nonce = "0" + nonce
	}

	hash := argon2.IDKey([]byte(blockHash), []byte(nonce), Argon2Iterations, Argon2Memory, Argon2Parallelism, Argon2KeyLength)

	var leadingBits []string
	for _, v := range hash {
		leadingBits = append(leadingBits, fmt.Sprintf("%08b", v))
	}

	joinStr := strings.Join(leadingBits, "")
	splitedStr := strings.Split(joinStr, "1")[0]

	// Computed difficulty: len of leading zero
	return int32(len(splitedStr))
}

func Mining(block types.MinerBlock, miningDifficulty int32, c *Client) (nonce uint64, difficulty int32, stop bool) {
	var latestNonceCounter uint64 = 0
	var nonceCounter uint64 = 0
	var latestTime time.Time = time.Now().UTC()
	var nowTime time.Time = latestTime

	var blockJSON []byte
	var latestHash string
	var nonceCounterDiff uint64
	var timeDiff time.Duration
	var workRate float64
	var computedDifficulty int32 = 0

	nonce = rand.Uint64()
	block.Nonce = fmt.Sprintf("%d", nonce)
	for {
		select {
		case <-c.StopMining:
			return 0, 0, true
		default:
		}
		if nonce == math.MaxInt32 {
			nonce = 0
		}
		nonce++
		block.Nonce = fmt.Sprintf("%d", nonce)

		if nonceCounter == math.MaxUint64 {
			latestNonceCounter = 0
			nonceCounter = 0
			nonceCounterDiff = 0
		}
		nonceCounter++

		blockJSON, _ = json.Marshal(block)
		latestHash = fmt.Sprintf("%x", sha256.Sum256(blockJSON))

		computedDifficulty = validate(latestHash, nonce, miningDifficulty)
		if computedDifficulty >= miningDifficulty {
			return nonce, computedDifficulty, false
		}

		nonceCounterDiff = nonceCounter - latestNonceCounter
		if nonceCounterDiff%100 == 0 {
			nowTime = time.Now().UTC()
			timeDiff = nowTime.Sub(latestTime)
			workRate = math.Floor(float64(nonceCounterDiff) / timeDiff.Seconds())
			log.Infof("%d works at %.1f [Work/s]", nonceCounterDiff, workRate)

			nonce = rand.Uint64()

			latestNonceCounter = nonceCounter
			latestTime = nowTime
		}
	}
}
