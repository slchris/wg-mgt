package service

import (
	"fmt"
	"time"

	"github.com/slchris/wg-mgt/internal/domain"
	"github.com/slchris/wg-mgt/internal/pkg/ssh"
	"github.com/slchris/wg-mgt/internal/repository"
)

// NodeService handles node business logic.
type NodeService struct {
	nodeRepo *repository.NodeRepository
}

// NewNodeService creates a new NodeService.
func NewNodeService(nodeRepo *repository.NodeRepository) *NodeService {
	return &NodeService{nodeRepo: nodeRepo}
}

// Create creates a new node.
func (s *NodeService) Create(node *domain.Node) error {
	// Set default WG interface if not provided
	if node.WGInterface == "" {
		node.WGInterface = "wg0"
	}

	// Set initial status
	node.Status = domain.NodeStatusUnknown

	// Create node in database first
	if err := s.nodeRepo.Create(node); err != nil {
		return err
	}

	// Try to fetch WireGuard config via SSH
	if node.SSHKey != "" {
		go s.fetchWGConfigAsync(node.ID)
	}

	return nil
}

// fetchWGConfigAsync fetches WireGuard configuration asynchronously after node creation.
func (s *NodeService) fetchWGConfigAsync(nodeID uint) {
	node, err := s.nodeRepo.GetByID(nodeID)
	if err != nil {
		return
	}

	client, err := ssh.NewClient(node.Host, node.SSHPort, node.SSHUser, node.SSHKey)
	if err != nil {
		return
	}

	// Try to get WireGuard status
	status, err := client.GetWireGuardStatus(node.WGInterface)
	if err != nil {
		// WireGuard might not be configured yet, that's ok
		return
	}

	// Update node with discovered WG config
	now := time.Now()
	node.Status = domain.NodeStatusOnline
	node.LastSeen = &now

	if status.Address != "" {
		node.WGAddress = status.Address
	}
	if status.PublicKey != "" {
		node.PublicKey = status.PublicKey
	}
	if status.ListenPort > 0 {
		node.WGPort = status.ListenPort
	}

	_ = s.nodeRepo.Update(node)
}

// GetByID retrieves a node by ID.
func (s *NodeService) GetByID(id uint) (*domain.Node, error) {
	return s.nodeRepo.GetByID(id)
}

// GetAll retrieves all nodes.
func (s *NodeService) GetAll() ([]domain.Node, error) {
	return s.nodeRepo.GetAll()
}

// Update updates a node.
func (s *NodeService) Update(node *domain.Node) error {
	return s.nodeRepo.Update(node)
}

// Delete deletes a node.
func (s *NodeService) Delete(id uint) error {
	return s.nodeRepo.Delete(id)
}

// GetWithPeers retrieves a node with its peers.
func (s *NodeService) GetWithPeers(id uint) (*domain.Node, error) {
	return s.nodeRepo.GetWithPeers(id)
}

// CheckStatus checks the connectivity status of a node via SSH and retrieves WireGuard info.
func (s *NodeService) CheckStatus(id uint) (*domain.Node, error) {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Try to connect via SSH and get WireGuard status
	client, err := ssh.NewClient(node.Host, node.SSHPort, node.SSHUser, node.SSHKey)
	if err != nil {
		// SSH key parse failed, mark as offline
		node.Status = domain.NodeStatusOffline
		_ = s.nodeRepo.Update(node)
		return node, fmt.Errorf("failed to create SSH client: %w", err)
	}

	// Try to get WireGuard status - this actually connects via SSH
	status, err := client.GetWireGuardStatus(node.WGInterface)
	if err != nil {
		// SSH connection or command failed
		node.Status = domain.NodeStatusOffline
		_ = s.nodeRepo.Update(node)
		return node, fmt.Errorf("failed to get WireGuard status: %w", err)
	}

	// Update node with WireGuard info
	node.Status = domain.NodeStatusOnline
	node.LastSeen = &now

	if status.Address != "" {
		node.WGAddress = status.Address
	}
	if status.PublicKey != "" {
		node.PublicKey = status.PublicKey
	}
	if status.ListenPort > 0 {
		node.WGPort = status.ListenPort
	}

	// Update node in database
	if updateErr := s.nodeRepo.Update(node); updateErr != nil {
		return nil, updateErr
	}

	return node, nil
}

// GetSSHClient creates an SSH client for a node.
func (s *NodeService) GetSSHClient(id uint) (*ssh.Client, error) {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return ssh.NewClient(node.Host, node.SSHPort, node.SSHUser, node.SSHKey)
}

// GetWireGuardStatus retrieves WireGuard status from a node via SSH.
func (s *NodeService) GetWireGuardStatus(id uint) (*ssh.WireGuardStatus, error) {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	client, err := ssh.NewClient(node.Host, node.SSHPort, node.SSHUser, node.SSHKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client: %w", err)
	}

	status, err := client.GetWireGuardStatus(node.WGInterface)
	if err != nil {
		return nil, fmt.Errorf("failed to get WireGuard status: %w", err)
	}

	// Update node status and WG info based on SSH response
	now := time.Now()
	node.Status = domain.NodeStatusOnline
	node.LastSeen = &now

	// Update WG address and public key if retrieved from remote
	if status.Address != "" && node.WGAddress != status.Address {
		node.WGAddress = status.Address
	}
	if status.PublicKey != "" && node.PublicKey != status.PublicKey {
		node.PublicKey = status.PublicKey
	}
	if status.ListenPort > 0 && node.WGPort != status.ListenPort {
		node.WGPort = status.ListenPort
	}

	_ = s.nodeRepo.Update(node)

	return status, nil
}

// GetNodeSystemInfo retrieves system information from a node via SSH.
func (s *NodeService) GetNodeSystemInfo(id uint) (map[string]string, error) {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	client, err := ssh.NewClient(node.Host, node.SSHPort, node.SSHUser, node.SSHKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client: %w", err)
	}

	info, err := client.GetSystemInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get system info: %w", err)
	}

	// Check WireGuard installation
	installed, version, _ := client.CheckWireGuardInstalled()
	info["wireguard_installed"] = fmt.Sprintf("%v", installed)
	if version != "" {
		info["wireguard_version"] = version
	}

	return info, nil
}

// InitializeRequest contains the parameters for initializing WireGuard on a node.
type InitializeRequest struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

// InitializeWireGuard initializes WireGuard on a node.
// This will install WireGuard if needed, generate keys, and create the config file.
func (s *NodeService) InitializeWireGuard(id uint, req *InitializeRequest) (*ssh.InitializeWireGuardResult, error) {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	client, err := ssh.NewClient(node.Host, node.SSHPort, node.SSHUser, node.SSHKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client: %w", err)
	}

	// Use defaults if not provided
	interfaceName := node.WGInterface
	if interfaceName == "" {
		interfaceName = "wg0"
	}

	address := req.Address
	if address == "" {
		// Use a default private IP if not provided
		address = "10.0.0.1/24"
	}

	port := req.Port
	if port == 0 {
		port = 51820
	}

	// Initialize WireGuard
	result, err := client.InitializeWireGuard(interfaceName, address, port)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize WireGuard: %w", err)
	}

	// Update node with the new config
	now := time.Now()
	node.Status = domain.NodeStatusOnline
	node.LastSeen = &now
	node.WGInterface = interfaceName
	node.WGAddress = result.Address
	node.WGPort = result.Port
	node.PublicKey = result.PublicKey

	if err := s.nodeRepo.Update(node); err != nil {
		return nil, fmt.Errorf("failed to update node: %w", err)
	}

	return result, nil
}

// SaveWireGuardConfig saves the current WireGuard runtime config to file on the node.
func (s *NodeService) SaveWireGuardConfig(id uint) error {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return err
	}

	client, err := ssh.NewClient(node.Host, node.SSHPort, node.SSHUser, node.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}

	interfaceName := node.WGInterface
	if interfaceName == "" {
		interfaceName = "wg0"
	}

	if err := client.SaveWireGuardConfig(interfaceName); err != nil {
		return fmt.Errorf("failed to save WireGuard config: %w", err)
	}

	return nil
}

// RestartWireGuard restarts the WireGuard interface on a node.
func (s *NodeService) RestartWireGuard(id uint) error {
	node, err := s.nodeRepo.GetByID(id)
	if err != nil {
		return err
	}

	client, err := ssh.NewClient(node.Host, node.SSHPort, node.SSHUser, node.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}

	interfaceName := node.WGInterface
	if interfaceName == "" {
		interfaceName = "wg0"
	}

	if err := client.RestartWireGuard(interfaceName); err != nil {
		return fmt.Errorf("failed to restart WireGuard: %w", err)
	}

	// Update node status
	now := time.Now()
	node.Status = domain.NodeStatusOnline
	node.LastSeen = &now
	_ = s.nodeRepo.Update(node)

	return nil
}
