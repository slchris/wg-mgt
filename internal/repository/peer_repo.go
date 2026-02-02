package repository

import (
	"github.com/slchris/wg-mgt/internal/domain"
	"gorm.io/gorm"
)

// PeerRepository handles database operations for peers.
type PeerRepository struct {
	db *gorm.DB
}

// NewPeerRepository creates a new PeerRepository.
func NewPeerRepository(db *gorm.DB) *PeerRepository {
	return &PeerRepository{db: db}
}

// Create creates a new peer.
func (r *PeerRepository) Create(peer *domain.Peer) error {
	return r.db.Create(peer).Error
}

// GetByID retrieves a peer by ID.
func (r *PeerRepository) GetByID(id uint) (*domain.Peer, error) {
	var peer domain.Peer
	if err := r.db.Preload("Node").First(&peer, id).Error; err != nil {
		return nil, err
	}
	return &peer, nil
}

// GetAll retrieves all peers.
func (r *PeerRepository) GetAll() ([]domain.Peer, error) {
	var peers []domain.Peer
	if err := r.db.Preload("Node").Find(&peers).Error; err != nil {
		return nil, err
	}
	return peers, nil
}

// GetByNodeID retrieves all peers for a node.
func (r *PeerRepository) GetByNodeID(nodeID uint) ([]domain.Peer, error) {
	var peers []domain.Peer
	if err := r.db.Preload("Node").Where("node_id = ?", nodeID).Find(&peers).Error; err != nil {
		return nil, err
	}
	return peers, nil
}

// Update updates a peer.
func (r *PeerRepository) Update(peer *domain.Peer) error {
	return r.db.Save(peer).Error
}

// Delete deletes a peer.
func (r *PeerRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Peer{}, id).Error
}
