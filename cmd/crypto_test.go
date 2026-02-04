package main

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	params := kdfParams{
		Time:    1,
		Memory:  8 * 1024,
		Threads: 1,
		KeyLen:  32,
		Salt:    []byte("1234567890abcdef"),
	}
	master := deriveKey("password", params)

	e, err := encryptEntry(master, "alpha", []byte("secret"))
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	got, err := decryptEntry(master, "alpha", e)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if string(got) != "secret" {
		t.Fatalf("expected secret, got %q", string(got))
	}
}

func TestDecryptWrongKey(t *testing.T) {
	params := kdfParams{
		Time:    1,
		Memory:  8 * 1024,
		Threads: 1,
		KeyLen:  32,
		Salt:    []byte("1234567890abcdef"),
	}
	master := deriveKey("password", params)

	e, err := encryptEntry(master, "alpha", []byte("secret"))
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	other := deriveKey("other", params)
	if _, err := decryptEntry(other, "alpha", e); err == nil {
		t.Fatalf("expected decrypt error with wrong password")
	}
}
