package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func IsValidAddress(addr string) bool {
	hexAddr, err := base64.StdEncoding.DecodeString(addr)
	if err != nil {
		return false
	}
	log.Debugf("Check validity of the address: %s (%s)", addr, hexAddr)

	version := hexAddr[:42]
	checksum := hexAddr[42:]
	doubleSha256 := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%x", sha256.Sum256(version)))))

	switch {
	case len(hexAddr) != 48:
		return false
	case bytes.Compare(checksum, []byte(doubleSha256)[:6]) != 0:
		return false
	}
	return true
}
