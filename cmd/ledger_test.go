package main

import (
	"path/filepath"
	"testing"
)

func TestLedgerRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ledger.encrypted")

	params := kdfParams{
		Time:    1,
		Memory:  8 * 1024,
		Threads: 1,
		KeyLen:  32,
		Salt:    []byte("1234567890abcdef"),
	}
	led := newLedger(params)

	master := deriveKey("password", params)
	e1, err := encryptEntry(master, "alpha", []byte("secret"))
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	led.Entries["alpha"] = e1

	if err := saveLedger(path, led); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := loadLedger(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(loaded.Entries))
	}

	e2, ok := loaded.Entries["alpha"]
	if !ok {
		t.Fatalf("missing key 'alpha'")
	}

	got, err := decryptEntry(master, "alpha", e2)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if string(got) != "secret" {
		t.Fatalf("expected secret, got %q", string(got))
	}
}
