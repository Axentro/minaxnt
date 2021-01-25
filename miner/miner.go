package miner

import (
	"crypto/sha256"
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
	argon2Memory      = 64 * 1024
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
	Nonce      string
	Difficulty int32
	Timestamp  int64
	Sha256Sum  string
}

func Mining(block types.MinerBlock, miningDifficulty int32, c *Client) (mr MiningResult, stop bool) {
	var blockJSON []byte
	var diff int32
	var blockHash [32]byte
	var blockHashString string

	nonce := rand.Uint64()
	block.Difficulty = miningDifficulty
	for {
		if c.StopClient.IsSet() {
			return MiningResult{}, true
		}
		block.Timestamp = time.Now().Unix()

		if nonce == math.MaxUint64 {
			nonce = 0
		} else if c.Stats.Counter()%100 == 0 {
			nonce = rand.Uint64()
		} else {
			nonce++
		}
		block.Nonce = strconv.FormatUint(nonce, 10)
		blockJSON, _ = json.Marshal(block)
		blockHash = sha256.Sum256(blockJSON)
		blockHashString = hex.EncodeToString(blockHash[:])
		diff = computeDifficulty(blockHashString, nonce)
		go c.Stats.Incr()
		if diff >= miningDifficulty {
			mr.Nonce = block.Nonce
			mr.Difficulty = block.Difficulty
			mr.Timestamp = block.Timestamp
			mr.Sha256Sum = blockHashString
			return mr, false
		}
	}
}
