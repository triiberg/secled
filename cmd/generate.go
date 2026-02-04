package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

func generateUUIDv4() (string, error) {
	b := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	hexStr := hex.EncodeToString(b)
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		hexStr[0:8],
		hexStr[8:12],
		hexStr[12:16],
		hexStr[16:20],
		hexStr[20:32],
	), nil
}

func generate256bSecret() (string, error) {
	b := make([]byte, 256)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
