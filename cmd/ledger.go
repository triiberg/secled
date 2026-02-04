package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

const (
	ledgerMagic   = "SECLED1"
	ledgerVersion = uint8(1)

	nonceSize = 12

	maxKeyLen   = 8 * 1024
	maxValueLen = 10 * 1024 * 1024
)

type kdfParams struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	Salt    []byte
}

type entry struct {
	Nonce      []byte
	Ciphertext []byte
}

type ledger struct {
	Params  kdfParams
	Entries map[string]entry
}

func ledgerPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(exe)
	return filepath.Join(dir, "ledger.encrypted"), nil
}

func newLedger(params kdfParams) *ledger {
	return &ledger{
		Params:  params,
		Entries: make(map[string]entry),
	}
}

func loadLedger(path string) (*ledger, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	magic := make([]byte, len(ledgerMagic))
	if _, err := io.ReadFull(f, magic); err != nil {
		return nil, err
	}
	if string(magic) != ledgerMagic {
		return nil, errors.New("invalid ledger header")
	}

	version, err := readUint8(f)
	if err != nil {
		return nil, err
	}
	if version != ledgerVersion {
		return nil, fmt.Errorf("unsupported ledger version: %d", version)
	}

	timeParam, err := readUint32(f)
	if err != nil {
		return nil, err
	}
	memParam, err := readUint32(f)
	if err != nil {
		return nil, err
	}
	threadsParam, err := readUint8(f)
	if err != nil {
		return nil, err
	}
	keyLen, err := readUint32(f)
	if err != nil {
		return nil, err
	}
	saltLen, err := readUint8(f)
	if err != nil {
		return nil, err
	}
	if saltLen == 0 {
		return nil, errors.New("invalid salt length")
	}

	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(f, salt); err != nil {
		return nil, err
	}

	count, err := readUint32(f)
	if err != nil {
		return nil, err
	}

	led := &ledger{
		Params: kdfParams{
			Time:    timeParam,
			Memory:  memParam,
			Threads: threadsParam,
			KeyLen:  keyLen,
			Salt:    salt,
		},
		Entries: make(map[string]entry),
	}

	for i := uint32(0); i < count; i++ {
		keyLen, err := readUint32(f)
		if err != nil {
			return nil, err
		}
		if keyLen == 0 || keyLen > maxKeyLen {
			return nil, errors.New("invalid key length")
		}
		keyBytes := make([]byte, keyLen)
		if _, err := io.ReadFull(f, keyBytes); err != nil {
			return nil, err
		}

		nonce := make([]byte, nonceSize)
		if _, err := io.ReadFull(f, nonce); err != nil {
			return nil, err
		}

		cipherLen, err := readUint32(f)
		if err != nil {
			return nil, err
		}
		if cipherLen == 0 || cipherLen > maxValueLen {
			return nil, errors.New("invalid ciphertext length")
		}
		cipherText := make([]byte, cipherLen)
		if _, err := io.ReadFull(f, cipherText); err != nil {
			return nil, err
		}

		led.Entries[string(keyBytes)] = entry{Nonce: nonce, Ciphertext: cipherText}
	}

	return led, nil
}

func saveLedger(path string, led *ledger) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "ledger.encrypted.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmp.Write([]byte(ledgerMagic)); err != nil {
		return err
	}
	if err := writeUint8(tmp, ledgerVersion); err != nil {
		return err
	}

	if err := writeUint32(tmp, led.Params.Time); err != nil {
		return err
	}
	if err := writeUint32(tmp, led.Params.Memory); err != nil {
		return err
	}
	if err := writeUint8(tmp, led.Params.Threads); err != nil {
		return err
	}
	if err := writeUint32(tmp, led.Params.KeyLen); err != nil {
		return err
	}
	if err := writeUint8(tmp, uint8(len(led.Params.Salt))); err != nil {
		return err
	}
	if _, err := tmp.Write(led.Params.Salt); err != nil {
		return err
	}

	keys := make([]string, 0, len(led.Entries))
	for k := range led.Entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if err := writeUint32(tmp, uint32(len(keys))); err != nil {
		return err
	}

	for _, key := range keys {
		e := led.Entries[key]
		if err := writeUint32(tmp, uint32(len(key))); err != nil {
			return err
		}
		if _, err := tmp.Write([]byte(key)); err != nil {
			return err
		}
		if len(e.Nonce) != nonceSize {
			return errors.New("invalid nonce size")
		}
		if _, err := tmp.Write(e.Nonce); err != nil {
			return err
		}
		if err := writeUint32(tmp, uint32(len(e.Ciphertext))); err != nil {
			return err
		}
		if _, err := tmp.Write(e.Ciphertext); err != nil {
			return err
		}
	}

	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		if runtime.GOOS == "windows" {
			_ = os.Remove(path)
			if err := os.Rename(tmpPath, path); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if runtime.GOOS != "windows" {
		_ = os.Chmod(path, 0o600)
	}

	return nil
}

func readUint8(r io.Reader) (uint8, error) {
	var b [1]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return b[0], nil
}

func readUint32(r io.Reader) (uint32, error) {
	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b[:]), nil
}

func writeUint8(w io.Writer, v uint8) error {
	_, err := w.Write([]byte{v})
	return err
}

func writeUint32(w io.Writer, v uint32) error {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	_, err := w.Write(b[:])
	return err
}
