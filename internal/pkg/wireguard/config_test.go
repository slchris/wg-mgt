package wireguard

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_GenerateServerConfig(t *testing.T) {
	config := &Config{
		Interface: InterfaceConfig{
			PrivateKey: "test-private-key",
			Address:    "10.0.0.1/24",
			ListenPort: 51820,
		},
		Peers: []PeerConfig{
			{
				PublicKey:           "peer-public-key",
				AllowedIPs:          []string{"10.0.0.2/32"},
				PersistentKeepalive: 25,
			},
		},
	}

	result := config.GenerateServerConfig()

	assert.Contains(t, result, "[Interface]")
	assert.Contains(t, result, "PrivateKey = test-private-key")
	assert.Contains(t, result, "Address = 10.0.0.1/24")
	assert.Contains(t, result, "ListenPort = 51820")
	assert.Contains(t, result, "[Peer]")
	assert.Contains(t, result, "PublicKey = peer-public-key")
	assert.Contains(t, result, "AllowedIPs = 10.0.0.2/32")
	assert.Contains(t, result, "PersistentKeepalive = 25")
}

func TestConfig_GenerateServerConfig_WithOptionalFields(t *testing.T) {
	config := &Config{
		Interface: InterfaceConfig{
			PrivateKey: "test-private-key",
			Address:    "10.0.0.1/24",
			ListenPort: 51820,
			MTU:        1420,
			PostUp:     "iptables -A FORWARD -i %i -j ACCEPT",
			PostDown:   "iptables -D FORWARD -i %i -j ACCEPT",
		},
		Peers: []PeerConfig{
			{
				PublicKey:    "peer-public-key",
				PresharedKey: "preshared-key",
				AllowedIPs:   []string{"10.0.0.2/32", "192.168.1.0/24"},
			},
		},
	}

	result := config.GenerateServerConfig()

	assert.Contains(t, result, "MTU = 1420")
	assert.Contains(t, result, "PostUp = iptables -A FORWARD -i %i -j ACCEPT")
	assert.Contains(t, result, "PostDown = iptables -D FORWARD -i %i -j ACCEPT")
	assert.Contains(t, result, "PresharedKey = preshared-key")
	assert.Contains(t, result, "AllowedIPs = 10.0.0.2/32, 192.168.1.0/24")
}

func TestConfig_GenerateServerConfig_NoPeers(t *testing.T) {
	config := &Config{
		Interface: InterfaceConfig{
			PrivateKey: "test-private-key",
			Address:    "10.0.0.1/24",
			ListenPort: 51820,
		},
		Peers: []PeerConfig{},
	}

	result := config.GenerateServerConfig()

	assert.Contains(t, result, "[Interface]")
	assert.NotContains(t, result, "[Peer]")
}

func TestConfig_GenerateServerConfig_MultiplePeers(t *testing.T) {
	config := &Config{
		Interface: InterfaceConfig{
			PrivateKey: "test-private-key",
			Address:    "10.0.0.1/24",
			ListenPort: 51820,
		},
		Peers: []PeerConfig{
			{PublicKey: "peer1-key", AllowedIPs: []string{"10.0.0.2/32"}},
			{PublicKey: "peer2-key", AllowedIPs: []string{"10.0.0.3/32"}},
			{PublicKey: "peer3-key", AllowedIPs: []string{"10.0.0.4/32"}},
		},
	}

	result := config.GenerateServerConfig()

	assert.Equal(t, 3, strings.Count(result, "[Peer]"))
	assert.Contains(t, result, "peer1-key")
	assert.Contains(t, result, "peer2-key")
	assert.Contains(t, result, "peer3-key")
}

func TestGenerateClientConfig(t *testing.T) {
	cfg := &ClientConfig{
		PrivateKey:          "client-private-key",
		Address:             "10.0.0.2/32",
		DNS:                 []string{"1.1.1.1", "8.8.8.8"},
		ServerPubKey:        "server-public-key",
		PresharedKey:        "preshared-key",
		Endpoint:            "example.com:51820",
		AllowedIPs:          []string{"0.0.0.0/0"},
		PersistentKeepalive: 25,
	}

	result := GenerateClientConfig(cfg)

	assert.Contains(t, result, "[Interface]")
	assert.Contains(t, result, "PrivateKey = client-private-key")
	assert.Contains(t, result, "Address = 10.0.0.2/32")
	assert.Contains(t, result, "DNS = 1.1.1.1, 8.8.8.8")
	assert.Contains(t, result, "[Peer]")
	assert.Contains(t, result, "PublicKey = server-public-key")
	assert.Contains(t, result, "PresharedKey = preshared-key")
	assert.Contains(t, result, "Endpoint = example.com:51820")
	assert.Contains(t, result, "AllowedIPs = 0.0.0.0/0")
	assert.Contains(t, result, "PersistentKeepalive = 25")
}

func TestGenerateClientConfig_MinimalConfig(t *testing.T) {
	cfg := &ClientConfig{
		PrivateKey:   "client-private-key",
		Address:      "10.0.0.2/32",
		ServerPubKey: "server-public-key",
		AllowedIPs:   []string{"0.0.0.0/0"},
	}

	result := GenerateClientConfig(cfg)

	assert.Contains(t, result, "PrivateKey = client-private-key")
	assert.Contains(t, result, "Address = 10.0.0.2/32")
	assert.Contains(t, result, "PublicKey = server-public-key")
	assert.NotContains(t, result, "DNS =")
	assert.NotContains(t, result, "PresharedKey =")
	assert.NotContains(t, result, "Endpoint =")
	assert.NotContains(t, result, "PersistentKeepalive =")
}
