package wireguard

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePrivateKey(t *testing.T) {
	key, err := GeneratePrivateKey()
	require.NoError(t, err)
	assert.NotEmpty(t, key)
	assert.True(t, ValidateKey(key))
}

func TestDerivePublicKey(t *testing.T) {
	privateKey, err := GeneratePrivateKey()
	require.NoError(t, err)

	publicKey, err := DerivePublicKey(privateKey)
	require.NoError(t, err)
	assert.NotEmpty(t, publicKey)
	assert.True(t, ValidateKey(publicKey))
}

func TestDerivePublicKey_InvalidKey(t *testing.T) {
	_, err := DerivePublicKey("invalid-key")
	assert.Error(t, err)
}

func TestGenerateKeyPair(t *testing.T) {
	privateKey, publicKey, err := GenerateKeyPair()
	require.NoError(t, err)
	assert.NotEmpty(t, privateKey)
	assert.NotEmpty(t, publicKey)
	assert.True(t, ValidateKey(privateKey))
	assert.True(t, ValidateKey(publicKey))
}

func TestGeneratePresharedKey(t *testing.T) {
	key, err := GeneratePresharedKey()
	require.NoError(t, err)
	assert.NotEmpty(t, key)
	assert.True(t, ValidateKey(key))
}

func TestValidateKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"valid key", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", true},
		{"invalid base64", "not-valid-base64!", false},
		{"too short", "AAAA", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateKey(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateKeyPair_Uniqueness(t *testing.T) {
	priv1, pub1, err := GenerateKeyPair()
	require.NoError(t, err)

	priv2, pub2, err := GenerateKeyPair()
	require.NoError(t, err)

	assert.NotEqual(t, priv1, priv2)
	assert.NotEqual(t, pub1, pub2)
}
