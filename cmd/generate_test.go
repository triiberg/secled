package main

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestGenerateUUIDv4Format(t *testing.T) {
	uuid, err := generateUUIDv4()
	if err != nil {
		t.Fatalf("generateUUIDv4 failed: %v", err)
	}
	if len(uuid) != 36 {
		t.Fatalf("expected length 36, got %d", len(uuid))
	}
	if uuid[8] != '-' || uuid[13] != '-' || uuid[18] != '-' || uuid[23] != '-' {
		t.Fatalf("invalid UUID format: %s", uuid)
	}
	if uuid[14] != '4' {
		t.Fatalf("expected version 4, got %c", uuid[14])
	}
	switch uuid[19] {
	case '8', '9', 'a', 'b':
	default:
		t.Fatalf("invalid variant: %c", uuid[19])
	}
	for i, r := range uuid {
		if r == '-' {
			continue
		}
		if !strings.ContainsRune("0123456789abcdef", r) {
			t.Fatalf("invalid hex char at %d: %q", i, r)
		}
	}
}

func TestGenerate64Hex(t *testing.T) {
	secret, err := generate64Hex()
	if err != nil {
		t.Fatalf("generate64Hex failed: %v", err)
	}
	data, err := hex.DecodeString(secret)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(data) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(data))
	}
}
