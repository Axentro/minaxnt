package miner

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"minaxnt/types"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Iterations  = 1
	argon2Memory      = 1 << 16 // 64 Mio
	argon2Parallelism = 1
	argon2KeyLength   = 512
)

func computeDifficulty(blockHash string, blockNonce uint64) int32 {
	nonce := strconv.FormatUint(blockNonce, 16)
	if utf8.RuneCountInString(nonce)%2 != 0 {
		nonce = "0" + nonce
	}
	hash := argon2.IDKey([]byte(blockHash), []byte(nonce), argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength)
	var leadingBits []string
	for _, v := range hash {
		leadingBits = append(leadingBits, fmt.Sprintf("%08b", v))
	}
	joinStr := strings.Join(leadingBits, "")
	splitedStr := strings.Split(joinStr, "1")[0]
	// Computed difficulty: len of leading zero
	return int32(len(splitedStr))
}

type MiningResult struct {
	Nonce           string
	Difficulty      int32
	Timestamp       int64
	BlockHashString string
}

func Mining(block types.MinerBlock, miningDifficulty int32, c *Client) (mr MiningResult, stop bool) {
	var blockJSON []byte
	var diff int32
	var blockHashString string

	nonce := rand.Uint64()
	block.Difficulty = miningDifficulty
	for {
		select {
		case <-c.StopMiningChan:
			return MiningResult{}, true
		default:
		}

		if nonce == math.MaxUint64 {
			nonce = 0
		} else if c.Stats.Counter()%100 == 0 {
			nonce = rand.Uint64()
		} else {
			nonce++
		}
		block.Nonce = strconv.FormatUint(nonce, 10)
		blockJSON, _ = json.Marshal(block)
		blockHashString = toArgon2HexString(blockJSON)
		diff = computeDifficulty(blockHashString, nonce)
		go c.Stats.Incr()
		if diff >= miningDifficulty {
			mr.Nonce = block.Nonce
			mr.Difficulty = block.Difficulty
			mr.Timestamp = time.Now().UTC().Unix() * time.Hour.Milliseconds()
			mr.BlockHashString = blockHashString
			return mr, false
		}
	}
}

func toArgon2HexString(str []byte) string {
	res := argon2.IDKey(str, []byte("AXENTRO_BLOCKCHAIN"), 1, 1<<4, 1, 32)
	return hex.EncodeToString(res)
}
