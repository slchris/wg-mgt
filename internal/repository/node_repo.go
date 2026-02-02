package repository

import (
	"github.com/slchris/wg-mgt/internal/domain"
	"gorm.io/gorm"
)

// NodeRepository handles database operations for nodes.
type NodeRepository struct {
	db *gorm.DB
}

// NewNodeRepository creates a new NodeRepository.
func NewNodeRepository(db *gorm.DB) *NodeRepository {
	return &NodeRepository{db: db}
}

// Create creates a new node.
func (r *NodeRepository) Create(node *domain.Node) error {
	return r.db.Create(node).Error
}

// GetByID retrieves a node by ID.
func (r *NodeRepository) GetByID(id uint) (*domain.Node, error) {
	var node domain.Node
	if err := r.db.First(&node, id).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

// GetAll retrieves all nodes.
func (r *NodeRepository) GetAll() ([]domain.Node, error) {
	var nodes []domain.Node
	if err := r.db.Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

// Update updates a node.
func (r *NodeRepository) Update(node *domain.Node) error {
	return r.db.Save(node).Error
}

// Delete deletes a node.
func (r *NodeRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Node{}, id).Error
}

// GetWithPeers retrieves a node with its peers.
func (r *NodeRepository) GetWithPeers(id uint) (*domain.Node, error) {
	var node domain.Node
	if err := r.db.Preload("Peers").First(&node, id).Error; err != nil {
		return nil, err
	}
	return &node, nil
}
