package wireguard

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/curve25519"
)

// GeneratePrivateKey generates a new WireGuard private key.
func GeneratePrivateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}

	// Clamp the key for curve25519
	key[0] &= 248
	key[31] &= 127
	key[31] |= 64

	return base64.StdEncoding.EncodeToString(key), nil
}

// DerivePublicKey derives a public key from a private key.
func DerivePublicKey(privateKey string) (string, error) {
	privBytes, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return "", err
	}

	var pubKey, privKey [32]byte
	copy(privKey[:], privBytes)

	curve25519.ScalarBaseMult(&pubKey, &privKey)

	return base64.StdEncoding.EncodeToString(pubKey[:]), nil
}

// GenerateKeyPair generates a new WireGuard key pair.
func GenerateKeyPair() (privateKey, publicKey string, err error) {
	privateKey, err = GeneratePrivateKey()
	if err != nil {
		return "", "", err
	}

	publicKey, err = DerivePublicKey(privateKey)
	if err != nil {
		return "", "", err
	}

	return privateKey, publicKey, nil
}

// GeneratePresharedKey generates a new preshared key.
func GeneratePresharedKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// ValidateKey checks if a key is valid base64 and 32 bytes.
func ValidateKey(key string) bool {
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return false
	}
	return len(decoded) == 32
}
