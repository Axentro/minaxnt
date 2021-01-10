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
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Iterations  = 1
	argon2Memory      = 64 * 1024
	argon2Parallelism = 1
	argon2KeyLength   = 512
)

func computeDifficulty(blockHash string, blockNonce uint32, difficulty int32) int32 {
	nonce := strconv.FormatUint(uint64(blockNonce), 16)
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

func Mining(block types.MinerBlock, miningDifficulty int32, c *Client) (nonce uint32, difficulty int32, stop bool) {
	var blockJSON []byte
	var cDiff int32
	var tempoHash [32]byte

	nonce = rand.Uint32()
	block.Nonce = strconv.Itoa(int(nonce))
	for {
		if c.StopClient.IsSet() {
			return 0, 0, true
		}

		if nonce == math.MaxUint32 {
			nonce = 0
		} else if c.Stats.Counter()%250 == 0 {
			nonce = rand.Uint32()
		} else {
			nonce++
		}
		block.Nonce = strconv.Itoa(int(nonce))
		blockJSON, _ = json.Marshal(block)
		tempoHash = sha256.Sum256(blockJSON)
		cDiff = computeDifficulty(hex.EncodeToString(tempoHash[:]), nonce, miningDifficulty)
		go c.Stats.Incr()

		if cDiff >= miningDifficulty {
			return nonce, cDiff, false
		}
	}
}
