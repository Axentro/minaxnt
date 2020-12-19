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

var argonParams = &types.Argon2Params{
	Memory:      64 * 1024,
	Iterations:  1,
	Parallelism: 1,
	KeyLength:   512,
}

func validate(blockHash string, blockNonce int, difficulty int32) int32 {
	nonce := strconv.FormatInt(int64(blockNonce), 16)
	if len([]rune(nonce))%2 != 0 {
		nonce = "0" + nonce
	}

	hash := argon2.IDKey([]byte(blockHash), []byte(nonce), argonParams.Iterations, argonParams.Memory, argonParams.Parallelism, argonParams.KeyLength)

	var leadingBits []string
	for _, v := range hash {
		leadingBits = append(leadingBits, fmt.Sprintf("%08b", v))
	}

	joinStr := strings.Join(leadingBits, "")
	splitedStr := strings.Split(joinStr, "1")[0]

	// Computed difficulty: len of leading zero
	return int32(len(splitedStr))
}

func Mining(block types.MinerBlock, miningDifficulty int32) types.MinerBlock {
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

	nonce := rand.Int()
	block.Nonce = fmt.Sprintf("%d", nonce)
	for {
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
		if computedDifficulty == miningDifficulty {
			return block
		}

		nonceCounterDiff = nonceCounter - latestNonceCounter
		if nonceCounterDiff%100 == 0 {
			nowTime = time.Now().UTC()
			timeDiff = nowTime.Sub(latestTime)
			workRate = math.Floor(float64(nonceCounterDiff) / timeDiff.Seconds())
			log.Infof("%d works, %.1f [Work/s]", nonceCounterDiff, workRate)

			nonce = rand.Int()

			latestNonceCounter = nonceCounter
			latestTime = nowTime
		}
	}
}
