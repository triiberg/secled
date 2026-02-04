package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	defaultKDFTime    uint32 = 3
	defaultKDFMemory  uint32 = 64 * 1024
	defaultKDFThreads uint8  = 4
	defaultKDFKeyLen  uint32 = 32

	defaultSaltSize = 16
)

func defaultKDFParams() (kdfParams, error) {
	salt := make([]byte, defaultSaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return kdfParams{}, err
	}
	return kdfParams{
		Time:    defaultKDFTime,
		Memory:  defaultKDFMemory,
		Threads: defaultKDFThreads,
		KeyLen:  defaultKDFKeyLen,
		Salt:    salt,
	}, nil
}

func deriveKey(password string, params kdfParams) []byte {
	return argon2.IDKey([]byte(password), params.Salt, params.Time, params.Memory, params.Threads, params.KeyLen)
}

func encryptEntry(masterKey []byte, key string, plaintext []byte) (entry, error) {
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return entry{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return entry{}, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return entry{}, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, []byte(key))
	return entry{Nonce: nonce, Ciphertext: ciphertext}, nil
}

func decryptEntry(masterKey []byte, key string, e entry) ([]byte, error) {
	if len(e.Nonce) != nonceSize {
		return nil, errors.New("invalid nonce size")
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := gcm.Open(nil, e.Nonce, e.Ciphertext, []byte(key))
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
