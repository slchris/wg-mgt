package service

import (
	"github.com/slchris/wg-mgt/internal/domain"
	"github.com/slchris/wg-mgt/internal/repository"
)

// NetworkService handles network business logic.
type NetworkService struct {
	networkRepo *repository.NetworkRepository
}

// NewNetworkService creates a new NetworkService.
func NewNetworkService(networkRepo *repository.NetworkRepository) *NetworkService {
	return &NetworkService{networkRepo: networkRepo}
}

// Create creates a new network.
func (s *NetworkService) Create(network *domain.Network) error {
	return s.networkRepo.Create(network)
}

// GetByID retrieves a network by ID.
func (s *NetworkService) GetByID(id uint) (*domain.Network, error) {
	return s.networkRepo.GetByID(id)
}

// GetAll retrieves all networks.
func (s *NetworkService) GetAll() ([]domain.Network, error) {
	return s.networkRepo.GetAll()
}

// Update updates a network.
func (s *NetworkService) Update(network *domain.Network) error {
	return s.networkRepo.Update(network)
}

// Delete deletes a network.
func (s *NetworkService) Delete(id uint) error {
	return s.networkRepo.Delete(id)
}
