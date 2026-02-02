package service

import (
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/slchris/wg-mgt/internal/domain"
	"github.com/slchris/wg-mgt/internal/pkg/ssh"
	"github.com/slchris/wg-mgt/internal/pkg/wireguard"
	"github.com/slchris/wg-mgt/internal/repository"
)

// PeerService handles peer business logic.
type PeerService struct {
	peerRepo *repository.PeerRepository
	nodeRepo *repository.NodeRepository
}

// NewPeerService creates a new PeerService.
func NewPeerService(peerRepo *repository.PeerRepository, nodeRepo *repository.NodeRepository) *PeerService {
	return &PeerService{
		peerRepo: peerRepo,
		nodeRepo: nodeRepo,
	}
}

// Create creates a new peer.
func (s *PeerService) Create(peer *domain.Peer) error {
	// Generate WireGuard keys if not provided
	if peer.PrivateKey == "" || peer.PublicKey == "" {
		privateKey, publicKey, err := wireguard.GenerateKeyPair()
		if err != nil {
			return err
		}
		peer.PrivateKey = privateKey
		peer.PublicKey = publicKey
	}

	// Generate preshared key if not provided
	if peer.PresharedKey == "" {
		psk, err := wireguard.GeneratePresharedKey()
		if err != nil {
			return err
		}
		peer.PresharedKey = psk
	}

	// Auto-assign IP if not provided
	if peer.Address == "" && peer.NodeID > 0 {
		nextIP, err := s.GetNextAvailableIP(peer.NodeID)
		if err != nil {
			return fmt.Errorf("failed to get next available IP: %w", err)
		}
		peer.Address = nextIP
	}

	// Create in database first
	if err := s.peerRepo.Create(peer); err != nil {
		return err
	}

	// Sync to WireGuard server
	if err := s.syncPeerToServer(peer); err != nil {
		log.Printf("Warning: failed to sync peer to server: %v", err)
		// Don't fail the create, just log the warning
		// The peer can be synced later manually
	}

	return nil
}

// GetByID retrieves a peer by ID.
func (s *PeerService) GetByID(id uint) (*domain.Peer, error) {
	return s.peerRepo.GetByID(id)
}

// GetAll retrieves all peers.
func (s *PeerService) GetAll() ([]domain.Peer, error) {
	return s.peerRepo.GetAll()
}

// GetByNodeID retrieves all peers for a node.
func (s *PeerService) GetByNodeID(nodeID uint) ([]domain.Peer, error) {
	return s.peerRepo.GetByNodeID(nodeID)
}

// Update updates a peer.
func (s *PeerService) Update(peer *domain.Peer) error {
	// Get existing peer to check for changes
	existing, err := s.peerRepo.GetByID(peer.ID)
	if err != nil {
		return err
	}

	// Update in database
	if err := s.peerRepo.Update(peer); err != nil {
		return err
	}

	// If enabled status changed, sync to server
	if existing.Enabled != peer.Enabled {
		if peer.Enabled {
			// Re-add peer to server
			if err := s.syncPeerToServer(peer); err != nil {
				log.Printf("Warning: failed to sync peer to server: %v", err)
			}
		} else {
			// Remove peer from server (disable)
			if err := s.removePeerFromServer(existing); err != nil {
				log.Printf("Warning: failed to remove peer from server: %v", err)
			}
		}
	}

	return nil
}

// Delete deletes a peer.
func (s *PeerService) Delete(id uint) error {
	// Get peer first to remove from server
	peer, err := s.peerRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Remove from WireGuard server
	if err := s.removePeerFromServer(peer); err != nil {
		log.Printf("Warning: failed to remove peer from server: %v", err)
		// Continue with database deletion
	}

	return s.peerRepo.Delete(id)
}

// GenerateClientConfig generates WireGuard client configuration for a peer.
func (s *PeerService) GenerateClientConfig(peerID uint) (string, error) {
	peer, err := s.peerRepo.GetByID(peerID)
	if err != nil {
		return "", err
	}

	node, err := s.nodeRepo.GetByID(peer.NodeID)
	if err != nil {
		return "", err
	}

	// Use peer's DNS if set, otherwise default to 1.1.1.1
	dns := []string{"1.1.1.1"}
	if peer.DNS != "" {
		dns = strings.Split(peer.DNS, ",")
		for i := range dns {
			dns[i] = strings.TrimSpace(dns[i])
		}
	}

	// Use peer's AllowedIPs if set, otherwise default to 0.0.0.0/0
	allowedIPs := []string{"0.0.0.0/0"}
	if peer.AllowedIPs != "" {
		allowedIPs = strings.Split(peer.AllowedIPs, ",")
		for i := range allowedIPs {
			allowedIPs[i] = strings.TrimSpace(allowedIPs[i])
		}
	}

	// Build endpoint: use node's Endpoint if set, otherwise construct from Host:WGPort
	endpoint := node.Endpoint
	if endpoint == "" && node.Host != "" {
		endpoint = fmt.Sprintf("%s:%d", node.Host, node.WGPort)
	}

	cfg := &wireguard.ClientConfig{
		PrivateKey:          peer.PrivateKey,
		Address:             peer.Address,
		DNS:                 dns,
		ServerPubKey:        node.PublicKey,
		PresharedKey:        peer.PresharedKey,
		Endpoint:            endpoint,
		AllowedIPs:          allowedIPs,
		PersistentKeepalive: 25,
	}

	return wireguard.GenerateClientConfig(cfg), nil
}

// GetNextAvailableIP returns the next available IP address for a node's subnet.
func (s *PeerService) GetNextAvailableIP(nodeID uint) (string, error) {
	node, err := s.nodeRepo.GetByID(nodeID)
	if err != nil {
		return "", fmt.Errorf("node not found: %w", err)
	}

	if node.WGAddress == "" {
		return "", fmt.Errorf("node has no WireGuard address configured")
	}

	// Parse node's WG address (e.g., "10.99.0.1/24")
	_, ipNet, err := net.ParseCIDR(node.WGAddress)
	if err != nil {
		return "", fmt.Errorf("invalid node WG address: %w", err)
	}

	// Get all existing peers for this node
	peers, err := s.peerRepo.GetByNodeID(nodeID)
	if err != nil {
		return "", err
	}

	// Collect all used IPs (including node's IP)
	usedIPs := make(map[string]bool)

	// Add node's IP
	nodeIP := strings.Split(node.WGAddress, "/")[0]
	usedIPs[nodeIP] = true

	// Add all peer IPs
	for _, peer := range peers {
		if peer.Address != "" {
			peerIP := strings.Split(peer.Address, "/")[0]
			usedIPs[peerIP] = true
		}
	}

	// Find next available IP in the subnet
	// Start from .2 (assuming .1 is the server)
	ip := ipNet.IP.Mask(ipNet.Mask)
	ones, bits := ipNet.Mask.Size()
	maxHosts := (1 << (bits - ones)) - 2 // Exclude network and broadcast

	for i := 2; i <= maxHosts; i++ {
		nextIP := incrementIP(ip, i)
		if !usedIPs[nextIP.String()] && ipNet.Contains(nextIP) {
			// Return with same mask as node
			mask := strings.Split(node.WGAddress, "/")[1]
			return fmt.Sprintf("%s/%s", nextIP.String(), mask), nil
		}
	}

	return "", fmt.Errorf("no available IP addresses in subnet")
}

// incrementIP increments an IP address by n
func incrementIP(ip net.IP, n int) net.IP {
	result := make(net.IP, len(ip))
	copy(result, ip)

	for i := len(result) - 1; i >= 0 && n > 0; i-- {
		sum := int(result[i]) + n
		result[i] = byte(sum % 256)
		n = sum / 256
	}
	return result
}

// GetUsedIPs returns all used IPs for a node
func (s *PeerService) GetUsedIPs(nodeID uint) ([]string, error) {
	node, err := s.nodeRepo.GetByID(nodeID)
	if err != nil {
		return nil, err
	}

	peers, err := s.peerRepo.GetByNodeID(nodeID)
	if err != nil {
		return nil, err
	}

	var ips []string
	if node.WGAddress != "" {
		ips = append(ips, strings.Split(node.WGAddress, "/")[0])
	}

	for _, peer := range peers {
		if peer.Address != "" {
			ips = append(ips, strings.Split(peer.Address, "/")[0])
		}
	}

	// Sort IPs
	sort.Slice(ips, func(i, j int) bool {
		return ipToInt(ips[i]) < ipToInt(ips[j])
	})

	return ips, nil
}

func ipToInt(ipStr string) uint32 {
	parts := strings.Split(ipStr, ".")
	if len(parts) != 4 {
		return 0
	}
	var result uint32
	for _, part := range parts {
		n, _ := strconv.Atoi(part)
		// #nosec G115 - IP octets are always 0-255, safe to convert
		result = result*256 + uint32(n)
	}
	return result
}

// syncPeerToServer adds a peer to the WireGuard server via SSH.
// This creates client config files and updates server config following wg.sh pattern.
func (s *PeerService) syncPeerToServer(peer *domain.Peer) error {
	node, err := s.nodeRepo.GetByID(peer.NodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	if node.SSHKey == "" {
		return fmt.Errorf("node has no SSH key configured")
	}

	log.Printf("Syncing peer %s (pubkey: %s) to node %s (%s)", peer.Name, peer.PublicKey, node.Name, node.Host)

	client, err := ssh.NewClient(node.Host, node.SSHPort, node.SSHUser, node.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}

	// Use peer's address as allowed IPs (the IP assigned to the peer)
	// Convert to /32 for single host (e.g., 10.99.0.3/24 -> 10.99.0.3/32)
	allowedIPs := peer.Address
	if allowedIPs == "" {
		return fmt.Errorf("peer has no address")
	}

	// Extract IP and use /32 for peer's allowed-ips on server side
	ipOnly := strings.Split(allowedIPs, "/")[0]
	serverAllowedIPs := ipOnly + "/32"

	// Build endpoint: use node's Endpoint if set, otherwise construct from Host:WGPort
	endpoint := node.Endpoint
	if endpoint == "" && node.Host != "" {
		endpoint = fmt.Sprintf("%s:%d", node.Host, node.WGPort)
	}

	// Parse DNS
	var dns []string
	if peer.DNS != "" {
		for _, d := range strings.Split(peer.DNS, ",") {
			dns = append(dns, strings.TrimSpace(d))
		}
	}
	if len(dns) == 0 {
		dns = []string{"1.1.1.1"}
	}

	// Create peer config for SSH client
	cfg := &ssh.PeerConfig{
		Name:         peer.Name,
		PublicKey:    peer.PublicKey,
		PrivateKey:   peer.PrivateKey,
		PresharedKey: peer.PresharedKey,
		AllowedIPs:   serverAllowedIPs,
		Address:      peer.Address,
		DNS:          dns,
		Endpoint:     endpoint,
		ServerPubKey: node.PublicKey,
	}

	log.Printf("Adding peer with allowed-ips: %s", serverAllowedIPs)

	if err := client.AddPeer(node.WGInterface, cfg); err != nil {
		return fmt.Errorf("failed to add peer to WireGuard: %w", err)
	}

	log.Printf("Successfully synced peer %s to node %s", peer.Name, node.Name)
	return nil
}

// removePeerFromServer removes a peer from the WireGuard server via SSH.
// This also removes client config files and key files following wg.sh pattern.
func (s *PeerService) removePeerFromServer(peer *domain.Peer) error {
	node, err := s.nodeRepo.GetByID(peer.NodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	if node.SSHKey == "" {
		return fmt.Errorf("node has no SSH key configured")
	}

	client, err := ssh.NewClient(node.Host, node.SSHPort, node.SSHUser, node.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}

	if err := client.RemovePeer(node.WGInterface, peer.PublicKey, peer.Name); err != nil {
		return fmt.Errorf("failed to remove peer from WireGuard: %w", err)
	}

	log.Printf("Successfully removed peer %s from node %s", peer.Name, node.Name)
	return nil
}

// SyncAllPeersToServer syncs all enabled peers for a node to the WireGuard server.
func (s *PeerService) SyncAllPeersToServer(nodeID uint) error {
	peers, err := s.peerRepo.GetByNodeID(nodeID)
	if err != nil {
		return err
	}

	var lastErr error
	for _, peer := range peers {
		if peer.Enabled {
			if err := s.syncPeerToServer(&peer); err != nil {
				log.Printf("Failed to sync peer %s: %v", peer.Name, err)
				lastErr = err
			}
		}
	}

	return lastErr
}
