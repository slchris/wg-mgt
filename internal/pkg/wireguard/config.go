package wireguard

import (
	"fmt"
	"strings"
)

// InterfaceConfig represents WireGuard interface configuration.
type InterfaceConfig struct {
	PrivateKey string
	Address    string
	ListenPort int
	MTU        int
	PostUp     string
	PostDown   string
}

// PeerConfig represents a WireGuard peer configuration.
type PeerConfig struct {
	PublicKey           string
	PresharedKey        string
	AllowedIPs          []string
	Endpoint            string
	PersistentKeepalive int
}

// Config represents a complete WireGuard configuration.
type Config struct {
	Interface InterfaceConfig
	Peers     []PeerConfig
}

// GenerateServerConfig generates the server-side WireGuard configuration.
func (c *Config) GenerateServerConfig() string {
	var sb strings.Builder

	sb.WriteString("[Interface]\n")
	sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", c.Interface.PrivateKey))
	sb.WriteString(fmt.Sprintf("Address = %s\n", c.Interface.Address))
	sb.WriteString(fmt.Sprintf("ListenPort = %d\n", c.Interface.ListenPort))

	if c.Interface.MTU > 0 {
		sb.WriteString(fmt.Sprintf("MTU = %d\n", c.Interface.MTU))
	}
	if c.Interface.PostUp != "" {
		sb.WriteString(fmt.Sprintf("PostUp = %s\n", c.Interface.PostUp))
	}
	if c.Interface.PostDown != "" {
		sb.WriteString(fmt.Sprintf("PostDown = %s\n", c.Interface.PostDown))
	}

	for _, peer := range c.Peers {
		sb.WriteString("\n[Peer]\n")
		sb.WriteString(fmt.Sprintf("PublicKey = %s\n", peer.PublicKey))

		if peer.PresharedKey != "" {
			sb.WriteString(fmt.Sprintf("PresharedKey = %s\n", peer.PresharedKey))
		}
		if len(peer.AllowedIPs) > 0 {
			sb.WriteString(fmt.Sprintf("AllowedIPs = %s\n", strings.Join(peer.AllowedIPs, ", ")))
		}
		if peer.PersistentKeepalive > 0 {
			sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", peer.PersistentKeepalive))
		}
	}

	return sb.String()
}

// ClientConfig represents client-side WireGuard configuration.
type ClientConfig struct {
	PrivateKey          string
	Address             string
	DNS                 []string
	ServerPubKey        string
	PresharedKey        string
	Endpoint            string
	AllowedIPs          []string
	PersistentKeepalive int
}

// GenerateClientConfig generates client-side WireGuard configuration.
func GenerateClientConfig(cfg *ClientConfig) string {
	var sb strings.Builder

	sb.WriteString("[Interface]\n")
	sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", cfg.PrivateKey))
	sb.WriteString(fmt.Sprintf("Address = %s\n", cfg.Address))

	if len(cfg.DNS) > 0 {
		sb.WriteString(fmt.Sprintf("DNS = %s\n", strings.Join(cfg.DNS, ", ")))
	}

	sb.WriteString("\n[Peer]\n")
	sb.WriteString(fmt.Sprintf("PublicKey = %s\n", cfg.ServerPubKey))

	if cfg.PresharedKey != "" {
		sb.WriteString(fmt.Sprintf("PresharedKey = %s\n", cfg.PresharedKey))
	}
	if cfg.Endpoint != "" {
		sb.WriteString(fmt.Sprintf("Endpoint = %s\n", cfg.Endpoint))
	}
	if len(cfg.AllowedIPs) > 0 {
		sb.WriteString(fmt.Sprintf("AllowedIPs = %s\n", strings.Join(cfg.AllowedIPs, ", ")))
	}
	if cfg.PersistentKeepalive > 0 {
		sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", cfg.PersistentKeepalive))
	}

	return sb.String()
}
