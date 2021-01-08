package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	log "github.com/sirupsen/logrus"
)

func IsValidAddress(addr string) bool {
	if addr == "" {
		return false
	}
	hexAddr, err := base64.StdEncoding.DecodeString(addr)
	if err != nil {
		return false
	}
	log.Debugf("Check validity of the address: %s (%s)", addr, hexAddr)

	versionHash := sha256.Sum256(hexAddr[:42])
	version := hex.EncodeToString(versionHash[:])

	dblVersionHash := sha256.Sum256([]byte(version))
	dblVersion := hex.EncodeToString(dblVersionHash[:])

	checksum := hexAddr[42:]

	switch {
	case len(hexAddr) != 48:
		return false
	case !bytes.Equal(checksum, []byte(dblVersion)[:6]):
		return false
	}
	return true
}
