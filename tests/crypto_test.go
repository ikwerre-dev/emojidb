package tests

import (
	"bytes"
	"testing"

	"github.com/ikwerre-dev/emojidb/crypto"
)

func TestCrypto(t *testing.T) {
	key := "secret"
	data := []byte("hello world")

	encrypted, err := crypto.Encrypt(data, key)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	decrypted, err := crypto.Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	if !bytes.Equal(data, decrypted) {
		t.Error("data mismatch")
	}
}

func TestEmojiEncoding(t *testing.T) {
	data := []byte{0, 1, 2, 255}
	encoded := crypto.EncodeToEmojis(data)

	decoded, err := crypto.DecodeFromEmojis(encoded)
	if err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if !bytes.Equal(data, decoded) {
		t.Errorf("expected %v, got %v", data, decoded)
	}
}
