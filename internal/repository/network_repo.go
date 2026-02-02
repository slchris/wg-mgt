package repository

import (
	"github.com/slchris/wg-mgt/internal/domain"
	"gorm.io/gorm"
)

// NetworkRepository handles database operations for networks.
type NetworkRepository struct {
	db *gorm.DB
}

// NewNetworkRepository creates a new NetworkRepository.
func NewNetworkRepository(db *gorm.DB) *NetworkRepository {
	return &NetworkRepository{db: db}
}

// Create creates a new network.
func (r *NetworkRepository) Create(network *domain.Network) error {
	return r.db.Create(network).Error
}

// GetByID retrieves a network by ID.
func (r *NetworkRepository) GetByID(id uint) (*domain.Network, error) {
	var network domain.Network
	if err := r.db.First(&network, id).Error; err != nil {
		return nil, err
	}
	return &network, nil
}

// GetAll retrieves all networks.
func (r *NetworkRepository) GetAll() ([]domain.Network, error) {
	var networks []domain.Network
	if err := r.db.Find(&networks).Error; err != nil {
		return nil, err
	}
	return networks, nil
}

// Update updates a network.
func (r *NetworkRepository) Update(network *domain.Network) error {
	return r.db.Save(network).Error
}

// Delete deletes a network.
func (r *NetworkRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Network{}, id).Error
}
