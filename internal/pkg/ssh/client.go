package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Client represents an SSH client connection.
type Client struct {
	config *ssh.ClientConfig
	host   string
	port   int
	user   string
}

// knownHostsMu protects concurrent access to the known_hosts file.
var knownHostsMu sync.Mutex

// knownHostsPath returns the path to the application's known_hosts file.
// It is stored alongside the working directory in a data/ subdirectory.
func knownHostsPath() string {
	if p := os.Getenv("WG_MGT_KNOWN_HOSTS"); p != "" {
		return p
	}
	return filepath.Join("data", "known_hosts")
}

// ensureKnownHostsFile creates the known_hosts file and its parent directory if they don't exist.
func ensureKnownHostsFile() (string, error) {
	khPath := knownHostsPath()
	if err := os.MkdirAll(filepath.Dir(khPath), 0700); err != nil {
		return "", fmt.Errorf("failed to create known_hosts directory: %w", err)
	}
	if _, err := os.Stat(khPath); errors.Is(err, os.ErrNotExist) {
		f, err := os.OpenFile(khPath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return "", fmt.Errorf("failed to create known_hosts file: %w", err)
		}
		_ = f.Close()
	}
	return khPath, nil
}

// hostKeyCallback returns a host key callback that implements Trust On First Use (TOFU).
// On first connection to a host, the key is accepted and stored in the known_hosts file.
// On subsequent connections, the stored key is verified; changed keys are rejected.
func hostKeyCallback() (ssh.HostKeyCallback, error) {
	khPath, err := ensureKnownHostsFile()
	if err != nil {
		return nil, err
	}

	cb, err := knownhosts.New(khPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load known_hosts: %w", err)
	}

	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := cb(hostname, remote, key)
		if err == nil {
			return nil // Host key matches
		}

		var keyErr *knownhosts.KeyError
		if errors.As(err, &keyErr) && len(keyErr.Want) > 0 {
			// Host key changed — reject to prevent MITM
			return fmt.Errorf("SSH host key mismatch for %s (possible MITM attack): %w", hostname, err)
		}

		// Unknown host — TOFU: accept and store the key
		knownHostsMu.Lock()
		defer knownHostsMu.Unlock()

		f, err := os.OpenFile(khPath, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to open known_hosts for writing: %w", err)
		}
		defer func() { _ = f.Close() }()

		line := knownhosts.Line([]string{knownhosts.Normalize(hostname)}, key)
		if _, err := fmt.Fprintln(f, line); err != nil {
			return fmt.Errorf("failed to write to known_hosts: %w", err)
		}
		return nil
	}, nil
}

// NewClient creates a new SSH client.
func NewClient(host string, port int, user string, privateKey string) (*Client, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	hkCallback, err := hostKeyCallback()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize host key verification: %w", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: hkCallback,
		Timeout:         10 * time.Second,
	}

	return &Client{
		config: config,
		host:   host,
		port:   port,
		user:   user,
	}, nil
}

// shellEscape escapes a string for use in a shell command.
func shellEscape(s string) string {
	if s == "" {
		return "''"
	}
	// For simple cases, we can just wrap in single quotes and escape existing single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// sudoPrefix returns "sudo " if user is not root, otherwise empty string.
func (c *Client) sudoPrefix() string {
	if c.user == "root" {
		return ""
	}
	return "sudo "
}

// Connect establishes an SSH connection.
func (c *Client) Connect() (*ssh.Client, error) {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	return ssh.Dial("tcp", addr, c.config)
}

// Validation patterns for common SSH command parameters.
var (
	validInterfaceName = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]{0,14}$`)
	validWGKey         = regexp.MustCompile(`^[A-Za-z0-9+/]{43}=$`)
	validCIDR          = regexp.MustCompile(`^[0-9a-fA-F.:]+/[0-9]{1,3}$`)
	validPeerName      = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,63}$`)
)

// ValidateInterfaceName checks that an interface name is safe to use in commands.
func ValidateInterfaceName(name string) error {
	if !validInterfaceName.MatchString(name) {
		return fmt.Errorf("invalid interface name: %q", name)
	}
	return nil
}

// ValidateWGKey checks that a WireGuard key is valid base64 (44 chars).
func ValidateWGKey(key string) error {
	if !validWGKey.MatchString(key) {
		return fmt.Errorf("invalid WireGuard key format")
	}
	return nil
}

// ValidateAllowedIPs checks that each CIDR in a comma-separated list is valid.
func ValidateAllowedIPs(ips string) error {
	for _, cidr := range strings.Split(ips, ",") {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			return fmt.Errorf("invalid CIDR %q: %w", cidr, err)
		}
	}
	return nil
}

// ValidatePeerName checks that a peer name is safe to use in file paths and commands.
func ValidatePeerName(name string) error {
	if !validPeerName.MatchString(name) {
		return fmt.Errorf("invalid peer name: %q", name)
	}
	return nil
}

// runCommand executes a command on the remote host.
// This method is unexported to ensure all command construction goes through
// the validated public methods of this package.
func (c *Client) runCommand(cmd string) (string, error) {
	client, err := c.Connect()
	if err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}
	defer func() { _ = client.Close() }()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer func() { _ = session.Close() }()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("command failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// WireGuardStatus represents the status of a WireGuard interface.
type WireGuardStatus struct {
	Interface  string       `json:"interface"`
	Address    string       `json:"address"`
	PublicKey  string       `json:"public_key"`
	ListenPort int          `json:"listen_port"`
	Peers      []PeerStatus `json:"peers"`
	IsRunning  bool         `json:"is_running"`
	TotalRx    int64        `json:"total_rx"`
	TotalTx    int64        `json:"total_tx"`
}

// PeerStatus represents the status of a WireGuard peer.
type PeerStatus struct {
	PublicKey           string    `json:"public_key"`
	Endpoint            string    `json:"endpoint"`
	AllowedIPs          []string  `json:"allowed_ips"`
	LatestHandshake     time.Time `json:"latest_handshake"`
	TransferRx          int64     `json:"transfer_rx"`
	TransferTx          int64     `json:"transfer_tx"`
	PersistentKeepalive int       `json:"persistent_keepalive"`
}

// GetWireGuardStatus retrieves the WireGuard interface status.
func (c *Client) GetWireGuardStatus(interfaceName string) (*WireGuardStatus, error) {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return nil, err
	}
	// Check if interface exists and get status
	output, err := c.runCommand(fmt.Sprintf("wg show %s 2>/dev/null || echo 'NOT_RUNNING'", shellEscape(interfaceName)))
	if err != nil {
		return nil, err
	}

	status := &WireGuardStatus{
		Interface: interfaceName,
		IsRunning: !strings.Contains(output, "NOT_RUNNING"),
		Peers:     []PeerStatus{},
	}

	if !status.IsRunning {
		return status, nil
	}

	// Get interface IP address using ip command
	if addrOutput, err := c.runCommand(fmt.Sprintf("ip -4 addr show %s 2>/dev/null | grep inet | awk '{print $2}'", shellEscape(interfaceName))); err == nil {
		status.Address = strings.TrimSpace(addrOutput)
	}

	// Parse wg show output
	lines := strings.Split(output, "\n")
	var currentPeer *PeerStatus

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "public key:") {
			status.PublicKey = strings.TrimSpace(strings.TrimPrefix(line, "public key:"))
		} else if strings.HasPrefix(line, "listen port:") {
			_, _ = fmt.Sscanf(strings.TrimPrefix(line, "listen port:"), "%d", &status.ListenPort)
		} else if strings.HasPrefix(line, "peer:") {
			if currentPeer != nil {
				status.Peers = append(status.Peers, *currentPeer)
			}
			currentPeer = &PeerStatus{
				PublicKey:  strings.TrimSpace(strings.TrimPrefix(line, "peer:")),
				AllowedIPs: []string{},
			}
		} else if currentPeer != nil {
			c.parseWgPeerLine(line, currentPeer)
		}
	}

	if currentPeer != nil {
		status.Peers = append(status.Peers, *currentPeer)
	}

	// Calculate totals
	for _, peer := range status.Peers {
		status.TotalRx += peer.TransferRx
		status.TotalTx += peer.TransferTx
	}

	return status, nil
}

// parseWgPeerLine parses a line containing peer information.
func (c *Client) parseWgPeerLine(line string, peer *PeerStatus) {
	if strings.HasPrefix(line, "endpoint:") {
		peer.Endpoint = strings.TrimSpace(strings.TrimPrefix(line, "endpoint:"))
	} else if strings.HasPrefix(line, "allowed ips:") {
		ips := strings.TrimSpace(strings.TrimPrefix(line, "allowed ips:"))
		peer.AllowedIPs = strings.Split(ips, ", ")
	} else if strings.HasPrefix(line, "latest handshake:") {
		// Parse handshake time (e.g., "1 minute, 30 seconds ago")
		// For simplicity, we'll just mark it as recent if present
		hsStr := strings.TrimSpace(strings.TrimPrefix(line, "latest handshake:"))
		if hsStr != "(none)" {
			peer.LatestHandshake = time.Now() // Simplified
		}
	} else if strings.HasPrefix(line, "transfer:") {
		// Parse transfer (e.g., "1.23 MiB received, 4.56 MiB sent")
		transferStr := strings.TrimPrefix(line, "transfer:")
		c.parseTransfer(transferStr, peer)
	} else if strings.HasPrefix(line, "persistent keepalive:") {
		kaStr := strings.TrimSpace(strings.TrimPrefix(line, "persistent keepalive:"))
		if kaStr != "off" {
			_, _ = fmt.Sscanf(kaStr, "every %d seconds", &peer.PersistentKeepalive)
		}
	}
}
func (c *Client) parseTransfer(transferStr string, peer *PeerStatus) {
	// Parse "1.23 MiB received, 4.56 MiB sent" or similar
	parts := strings.Split(transferStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "received") {
			peer.TransferRx = c.parseSize(part)
		} else if strings.Contains(part, "sent") {
			peer.TransferTx = c.parseSize(part)
		}
	}
}

func (c *Client) parseSize(sizeStr string) int64 {
	var value float64
	var unit string
	_, _ = fmt.Sscanf(sizeStr, "%f %s", &value, &unit)

	multiplier := int64(1)
	switch {
	case strings.HasPrefix(unit, "KiB"):
		multiplier = 1024
	case strings.HasPrefix(unit, "MiB"):
		multiplier = 1024 * 1024
	case strings.HasPrefix(unit, "GiB"):
		multiplier = 1024 * 1024 * 1024
	case strings.HasPrefix(unit, "TiB"):
		multiplier = 1024 * 1024 * 1024 * 1024
	}

	return int64(value * float64(multiplier))
}

// CheckWireGuardInstalled checks if WireGuard is installed on the remote host.
func (c *Client) CheckWireGuardInstalled() (bool, string, error) {
	output, err := c.runCommand("which wg && wg --version 2>/dev/null || echo 'NOT_INSTALLED'")
	if err != nil {
		return false, "", err
	}

	if strings.Contains(output, "NOT_INSTALLED") {
		return false, "", nil
	}

	// Extract version from output
	lines := strings.Split(output, "\n")
	version := ""
	for _, line := range lines {
		if strings.Contains(line, "wireguard-tools") {
			version = strings.TrimSpace(line)
			break
		}
	}

	return true, version, nil
}

// GetSystemInfo retrieves basic system information.
func (c *Client) GetSystemInfo() (map[string]string, error) {
	info := make(map[string]string)

	// Get hostname
	if hostname, err := c.runCommand("hostname"); err == nil {
		info["hostname"] = strings.TrimSpace(hostname)
	}

	// Get OS info
	if osInfo, err := c.runCommand("cat /etc/os-release 2>/dev/null | grep PRETTY_NAME | cut -d'\"' -f2"); err == nil {
		info["os"] = strings.TrimSpace(osInfo)
	}

	// Get kernel version
	if kernel, err := c.runCommand("uname -r"); err == nil {
		info["kernel"] = strings.TrimSpace(kernel)
	}

	// Get uptime
	if uptime, err := c.runCommand("uptime -p 2>/dev/null || uptime"); err == nil {
		info["uptime"] = strings.TrimSpace(uptime)
	}

	return info, nil
}

// PeerConfig contains all the information needed to add a peer
type PeerConfig struct {
	Name         string // Peer name (used for config file name)
	PublicKey    string
	PrivateKey   string // Client's private key (for client config)
	PresharedKey string
	AllowedIPs   string   // Peer's IP (e.g., 10.99.0.3/32 for server side)
	Address      string   // Client's address (e.g., 10.99.0.3/24)
	DNS          []string // DNS servers for client config
	Endpoint     string   // Server endpoint for client config (e.g., server.com:51820)
	ServerPubKey string   // Server's public key (for client config)
}

// AddPeer adds a peer to the WireGuard interface and creates config files.
// This follows the wg.sh pattern: creates client config in /etc/wireguard/clients/,
// saves keys in /etc/wireguard/keys/, and appends [Peer] to server config.
func (c *Client) AddPeer(interfaceName string, cfg *PeerConfig) error {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return err
	}
	if err := ValidateWGKey(cfg.PublicKey); err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}
	if err := ValidateAllowedIPs(cfg.AllowedIPs); err != nil {
		return err
	}
	if err := ValidatePeerName(cfg.Name); err != nil {
		return err
	}
	sudo := c.sudoPrefix()
	wgDir := "/etc/wireguard"
	clientsDir := fmt.Sprintf("%s/clients", wgDir)
	keysDir := fmt.Sprintf("%s/keys", wgDir)
	confFile := fmt.Sprintf("%s/%s.conf", wgDir, shellEscape(interfaceName))

	// Ensure directories exist
	mkdirCmd := fmt.Sprintf("%smkdir -p %s %s", sudo, shellEscape(clientsDir), shellEscape(keysDir))
	if _, err := c.runCommand(mkdirCmd); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Save keys to files
	keysCmds := fmt.Sprintf(`
%scat > %s/%s_private.key << 'KEYEOF'
%s
KEYEOF
%scat > %s/%s_public.key << 'KEYEOF'
%s
KEYEOF
%scat > %s/%s_psk.key << 'KEYEOF'
%s
KEYEOF
%schmod 600 %s/%s_*.key
`, sudo, shellEscape(keysDir), shellEscape(cfg.Name), cfg.PrivateKey,
		sudo, shellEscape(keysDir), shellEscape(cfg.Name), cfg.PublicKey,
		sudo, shellEscape(keysDir), shellEscape(cfg.Name), cfg.PresharedKey,
		sudo, shellEscape(keysDir), shellEscape(cfg.Name))

	if _, err := c.runCommand(keysCmds); err != nil {
		return fmt.Errorf("failed to save keys: %w", err)
	}

	// Generate client config file
	dnsStr := "1.1.1.1"
	if len(cfg.DNS) > 0 {
		dnsStr = strings.Join(cfg.DNS, ", ")
	}

	clientConf := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
DNS = %s

[Peer]
PublicKey = %s
PresharedKey = %s
Endpoint = %s
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
`, cfg.PrivateKey, cfg.Address, dnsStr, cfg.ServerPubKey, cfg.PresharedKey, cfg.Endpoint)

	// Save client config
	clientConfCmd := fmt.Sprintf(`%scat > %s/%s.conf << 'CONFEOF'
%sCONFEOF
%schmod 600 %s/%s.conf
`, sudo, shellEscape(clientsDir), shellEscape(cfg.Name), clientConf, sudo, shellEscape(clientsDir), shellEscape(cfg.Name))

	if _, err := c.runCommand(clientConfCmd); err != nil {
		return fmt.Errorf("failed to save client config: %w", err)
	}

	// Append [Peer] section to server config
	peerSection := fmt.Sprintf(`
### Client: %s
[Peer]
PublicKey = %s
PresharedKey = %s
AllowedIPs = %s
`, cfg.Name, cfg.PublicKey, cfg.PresharedKey, cfg.AllowedIPs)

	appendCmd := fmt.Sprintf(`%scat >> %s << 'PEEREOF'
%sPEEREOF
`, sudo, confFile, peerSection)

	if _, err := c.runCommand(appendCmd); err != nil {
		return fmt.Errorf("failed to append peer to server config: %w", err)
	}

	// Apply peer dynamically using wg set
	var setCmd string
	if cfg.PresharedKey != "" {
		setCmd = fmt.Sprintf(`
PSK=$(mktemp)
echo '%s' > "$PSK"
%swg set %s peer %s preshared-key "$PSK" allowed-ips %s
rm -f "$PSK"
`, cfg.PresharedKey, sudo, shellEscape(interfaceName), shellEscape(cfg.PublicKey), shellEscape(cfg.AllowedIPs))
	} else {
		setCmd = fmt.Sprintf("%swg set %s peer %s allowed-ips %s", sudo, shellEscape(interfaceName), shellEscape(cfg.PublicKey), shellEscape(cfg.AllowedIPs))
	}

	if _, err := c.runCommand(setCmd); err != nil {
		return fmt.Errorf("failed to add peer dynamically: %w", err)
	}

	return nil
}

// AddPeerSimple is a simplified version for backward compatibility
func (c *Client) AddPeerSimple(interfaceName, publicKey, presharedKey, allowedIPs string) error {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return err
	}
	if err := ValidateWGKey(publicKey); err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}
	if err := ValidateAllowedIPs(allowedIPs); err != nil {
		return err
	}
	sudo := c.sudoPrefix()
	var cmd string
	if presharedKey != "" {
		cmd = fmt.Sprintf(`
PSK=$(mktemp)
echo '%s' > "$PSK"
%swg set %s peer %s preshared-key "$PSK" allowed-ips %s
rm -f "$PSK"
`, presharedKey, sudo, shellEscape(interfaceName), shellEscape(publicKey), shellEscape(allowedIPs))
	} else {
		cmd = fmt.Sprintf("%swg set %s peer %s allowed-ips %s", sudo, shellEscape(interfaceName), shellEscape(publicKey), shellEscape(allowedIPs))
	}

	if _, err := c.runCommand(cmd); err != nil {
		return fmt.Errorf("failed to add peer: %w", err)
	}

	return nil
}

// RemovePeer removes a peer from the WireGuard interface and cleans up config files.
// This follows the wg.sh pattern: uses awk to remove [Peer] section from server config,
// and removes client config and key files.
func (c *Client) RemovePeer(interfaceName, publicKey, peerName string) error {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return err
	}
	if err := ValidateWGKey(publicKey); err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}
	if peerName != "" {
		if err := ValidatePeerName(peerName); err != nil {
			return err
		}
	}
	sudo := c.sudoPrefix()
	wgDir := "/etc/wireguard"
	confFile := fmt.Sprintf("%s/%s.conf", wgDir, shellEscape(interfaceName))
	clientsDir := fmt.Sprintf("%s/clients", wgDir)
	keysDir := fmt.Sprintf("%s/keys", wgDir)

	// Remove peer dynamically
	rmCmd := fmt.Sprintf("%swg set %s peer %s remove", sudo, shellEscape(interfaceName), shellEscape(publicKey))
	if _, err := c.runCommand(rmCmd); err != nil {
		return fmt.Errorf("failed to remove peer dynamically: %w", err)
	}

	// Remove [Peer] section from server config using awk (wg.sh pattern)
	// Use a reliable approach with a temp file
	cleanupCmd := fmt.Sprintf(`
# Create backup
%scp %s %s.bak

# Remove peer section from config (method from wg.sh)
%sawk -v pk="%s" '
BEGIN {skip=0}
/^\[Peer\]/ {
    peer_start = NR
    peer_lines = $0
    while ((getline line) > 0) {
        if (line ~ /^\[/ || line ~ /^###/ || line == "") {
            if (peer_lines !~ pk) {
                print peer_lines
            }
            if (line != "") print line
            break
        }
        peer_lines = peer_lines "\n" line
    }
    next
}
/^### Client: / {
    comment = $0
    getline
    if ($0 ~ pk) {
        # Skip this comment and continue skipping peer block
        while ((getline) > 0 && $0 !~ /^(\[|###|$)/) {}
        if ($0 ~ /^(\[|###)/) print
        next
    }
    print comment
}
{print}
' %s > %s.new && %smv %s.new %s
`, sudo, confFile, confFile,
		sudo, publicKey, confFile, confFile, sudo, confFile, confFile)

	// Don't fail if config cleanup fails, peer is already removed dynamically
	_, _ = c.runCommand(cleanupCmd)

	// Remove client config file
	if peerName != "" {
		rmFilesCmd := fmt.Sprintf("%srm -f %s/%s.conf %s/%s_private.key %s/%s_public.key %s/%s_psk.key 2>/dev/null || true",
			sudo, shellEscape(clientsDir), shellEscape(peerName), shellEscape(keysDir), shellEscape(peerName), shellEscape(keysDir), shellEscape(peerName), shellEscape(keysDir), shellEscape(peerName))
		_, _ = c.runCommand(rmFilesCmd)
	}

	return nil
}

// RemovePeerSimple removes a peer with minimal cleanup (backward compatibility)
func (c *Client) RemovePeerSimple(interfaceName, publicKey string) error {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return err
	}
	if err := ValidateWGKey(publicKey); err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}
	sudo := c.sudoPrefix()
	cmd := fmt.Sprintf("%swg set %s peer %s remove", sudo, shellEscape(interfaceName), shellEscape(publicKey))
	_, err := c.runCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to remove peer: %w", err)
	}
	return nil
}

// UpdatePeerAllowedIPs updates the allowed IPs for a peer.
func (c *Client) UpdatePeerAllowedIPs(interfaceName, publicKey, allowedIPs string) error {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return err
	}
	if err := ValidateWGKey(publicKey); err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}
	if err := ValidateAllowedIPs(allowedIPs); err != nil {
		return err
	}
	sudo := c.sudoPrefix()
	cmd := fmt.Sprintf("%swg set %s peer %s allowed-ips %s", sudo, shellEscape(interfaceName), shellEscape(publicKey), shellEscape(allowedIPs))
	_, err := c.runCommand(cmd)
	if err != nil {
		return fmt.Errorf("failed to update peer: %w", err)
	}

	return nil
}

// GetClientConfig retrieves the client configuration from the server
func (c *Client) GetClientConfig(peerName string) (string, error) {
	if err := ValidatePeerName(peerName); err != nil {
		return "", err
	}
	sudo := c.sudoPrefix()
	cmd := fmt.Sprintf("%scat /etc/wireguard/clients/%s.conf 2>/dev/null", sudo, shellEscape(peerName))
	output, err := c.runCommand(cmd)
	if err != nil {
		return "", fmt.Errorf("client config not found: %w", err)
	}
	return strings.TrimSpace(output), nil
}

// ListClients lists all client config files on the server
func (c *Client) ListClients() ([]string, error) {
	sudo := c.sudoPrefix()
	cmd := fmt.Sprintf("%sls /etc/wireguard/clients/*.conf 2>/dev/null | xargs -n1 basename 2>/dev/null | sed 's/.conf$//' || true", sudo)
	output, err := c.runCommand(cmd)
	if err != nil {
		return nil, err
	}

	var clients []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line != "" {
			clients = append(clients, line)
		}
	}
	return clients, nil
}

// PeerExists checks if a peer exists on the WireGuard interface.
func (c *Client) PeerExists(interfaceName, publicKey string) (bool, error) {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return false, err
	}
	if err := ValidateWGKey(publicKey); err != nil {
		return false, fmt.Errorf("invalid public key: %w", err)
	}
	sudo := c.sudoPrefix()
	cmd := fmt.Sprintf("%swg show %s peers 2>/dev/null | grep -q '%s' && echo 'EXISTS' || echo 'NOT_EXISTS'", sudo, shellEscape(interfaceName), shellEscape(publicKey))
	output, err := c.runCommand(cmd)
	if err != nil {
		return false, err
	}
	return strings.Contains(output, "EXISTS"), nil
}

// InitializeWireGuardResult contains the result of WireGuard initialization
type InitializeWireGuardResult struct {
	Installed     bool   `json:"installed"`
	WasInstalled  bool   `json:"was_installed"` // true if we just installed it
	Configured    bool   `json:"configured"`
	WasConfigured bool   `json:"was_configured"` // true if we just configured it
	Interface     string `json:"interface"`
	Address       string `json:"address"`
	Port          int    `json:"port"`
	PublicKey     string `json:"public_key"`
	PrivateKey    string `json:"private_key"`
	Message       string `json:"message"`
}

// InstallWireGuard installs WireGuard on the remote host
func (c *Client) InstallWireGuard() error {
	sudo := c.sudoPrefix()

	// Detect package manager and install
	installCmd := fmt.Sprintf(`
%scommand -v apt-get >/dev/null 2>&1 && { %sapt-get update && %sapt-get install -y wireguard; exit $?; }
%scommand -v yum >/dev/null 2>&1 && { %syum install -y epel-release && %syum install -y wireguard-tools; exit $?; }
%scommand -v dnf >/dev/null 2>&1 && { %sdnf install -y wireguard-tools; exit $?; }
%scommand -v pacman >/dev/null 2>&1 && { %spacman -Sy --noconfirm wireguard-tools; exit $?; }
%scommand -v apk >/dev/null 2>&1 && { %sapk add wireguard-tools; exit $?; }
echo "Unsupported package manager" && exit 1
`, sudo, sudo, sudo, sudo, sudo, sudo, sudo, sudo, sudo, sudo, sudo, sudo)

	_, err := c.runCommand(installCmd)
	return err
}

// InitializeWireGuard initializes WireGuard on the remote host
// This includes: installing if needed, generating keys, creating config file, starting service
func (c *Client) InitializeWireGuard(interfaceName, address string, port int) (*InitializeWireGuardResult, error) {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return nil, err
	}
	sudo := c.sudoPrefix()
	result := &InitializeWireGuardResult{
		Interface: interfaceName,
		Address:   address,
		Port:      port,
	}

	// Check if WireGuard is installed
	installed, _, err := c.CheckWireGuardInstalled()
	if err != nil {
		return nil, fmt.Errorf("failed to check WireGuard installation: %w", err)
	}
	result.Installed = installed

	// Install if not installed
	if !installed {
		if err := c.InstallWireGuard(); err != nil {
			return nil, fmt.Errorf("failed to install WireGuard: %w", err)
		}
		result.Installed = true
		result.WasInstalled = true
	}

	// Check if config already exists
	confPath := fmt.Sprintf("/etc/wireguard/%s.conf", shellEscape(interfaceName))
	checkCmd := fmt.Sprintf("%stest -f %s && echo 'EXISTS' || echo 'NOT_EXISTS'", sudo, confPath)
	output, err := c.runCommand(checkCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to check config: %w", err)
	}
	configExists := strings.Contains(output, "EXISTS")

	if configExists {
		// Config exists, just get the info
		result.Configured = true
		status, err := c.GetWireGuardStatus(interfaceName)
		if err == nil && status.IsRunning {
			result.PublicKey = status.PublicKey
			result.Address = status.Address
			result.Port = status.ListenPort
		} else {
			// Try to read public key from config
			pubKeyCmd := fmt.Sprintf("%sgrep -o 'PrivateKey = .*' %s | cut -d' ' -f3 | wg pubkey", sudo, confPath)
			if pubKey, err := c.runCommand(pubKeyCmd); err == nil {
				result.PublicKey = strings.TrimSpace(pubKey)
			}
		}
		result.Message = "WireGuard already configured"
		return result, nil
	}

	// Generate keys
	keyGenCmd := fmt.Sprintf(`
%smkdir -p /etc/wireguard/keys /etc/wireguard/clients
PRIVATE_KEY=$(%swg genkey)
PUBLIC_KEY=$(echo "$PRIVATE_KEY" | %swg pubkey)
echo "$PRIVATE_KEY" | %stee /etc/wireguard/keys/%s_private.key > /dev/null
echo "$PUBLIC_KEY" | %stee /etc/wireguard/keys/%s_public.key > /dev/null
%schmod 600 /etc/wireguard/keys/*
echo "$PRIVATE_KEY"
echo "$PUBLIC_KEY"
`, sudo, sudo, sudo, sudo, shellEscape(interfaceName), sudo, shellEscape(interfaceName), sudo)

	keysOutput, err := c.runCommand(keyGenCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	keys := strings.Split(strings.TrimSpace(keysOutput), "\n")
	if len(keys) >= 2 {
		result.PrivateKey = strings.TrimSpace(keys[0])
		result.PublicKey = strings.TrimSpace(keys[1])
	}

	// Create config file
	configContent := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
ListenPort = %d
SaveConfig = false
`, result.PrivateKey, address, port)

	createConfigCmd := fmt.Sprintf(`%scat > %s << 'WGCONF'
%sWGCONF
%schmod 600 %s
`, sudo, confPath, configContent, sudo, confPath)

	if _, err := c.runCommand(createConfigCmd); err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	// Enable IP forwarding
	forwardCmd := fmt.Sprintf(`
%ssed -i '/net.ipv4.ip_forward/d' /etc/sysctl.conf
echo 'net.ipv4.ip_forward = 1' | %stee -a /etc/sysctl.conf
%ssysctl -p
`, sudo, sudo, sudo)
	_, _ = c.runCommand(forwardCmd)

	// Start WireGuard
	startCmd := fmt.Sprintf(`
%ssystemctl enable wg-quick@%s 2>/dev/null || true
%ssystemctl start wg-quick@%s 2>/dev/null || %swg-quick up %s
`, sudo, shellEscape(interfaceName), sudo, shellEscape(interfaceName), sudo, shellEscape(interfaceName))

	if _, err := c.runCommand(startCmd); err != nil {
		return nil, fmt.Errorf("failed to start WireGuard: %w", err)
	}

	result.Configured = true
	result.WasConfigured = true
	result.Message = "WireGuard initialized successfully"

	return result, nil
}

// SaveWireGuardConfig saves the current WireGuard runtime config to file
func (c *Client) SaveWireGuardConfig(interfaceName string) error {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return err
	}
	sudo := c.sudoPrefix()

	// Get current config from wg show
	showCmd := fmt.Sprintf("%swg showconf %s", sudo, shellEscape(interfaceName))
	config, err := c.runCommand(showCmd)
	if err != nil {
		return fmt.Errorf("failed to get current config: %w", err)
	}

	// Get interface address
	addrCmd := fmt.Sprintf("%sip -4 addr show %s | grep inet | awk '{print $2}'", sudo, shellEscape(interfaceName))
	addr, _ := c.runCommand(addrCmd)
	addr = strings.TrimSpace(addr)

	// Merge Address into config (wg showconf doesn't include it)
	if addr != "" && !strings.Contains(config, "Address") {
		// Insert Address after [Interface]
		config = strings.Replace(config, "[Interface]\n", fmt.Sprintf("[Interface]\nAddress = %s\n", addr), 1)
	}

	// Add SaveConfig = false
	if !strings.Contains(config, "SaveConfig") {
		config = strings.Replace(config, "[Interface]\n", "[Interface]\nSaveConfig = false\n", 1)
	}

	// Write to config file
	confPath := fmt.Sprintf("/etc/wireguard/%s.conf", shellEscape(interfaceName))
	writeCmd := fmt.Sprintf(`%scat > %s << 'WGCONF'
%sWGCONF
%schmod 600 %s
`, sudo, confPath, config, sudo, confPath)

	if _, err := c.runCommand(writeCmd); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// RestartWireGuard restarts the WireGuard interface
func (c *Client) RestartWireGuard(interfaceName string) error {
	if err := ValidateInterfaceName(interfaceName); err != nil {
		return err
	}
	sudo := c.sudoPrefix()
	cmd := fmt.Sprintf(`%swg-quick down %s 2>/dev/null; %swg-quick up %s`, sudo, shellEscape(interfaceName), sudo, shellEscape(interfaceName))
	_, err := c.runCommand(cmd)
	return err
}
