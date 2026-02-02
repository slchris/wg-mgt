package domain

import "time"

// Peer represents a WireGuard peer (client).
type Peer struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"size:100;not null" json:"name"`
	NodeID       uint       `gorm:"not null" json:"node_id"`
	Node         *Node      `gorm:"foreignKey:NodeID" json:"node,omitempty"`
	PublicKey    string     `gorm:"size:64;not null" json:"public_key"`
	PrivateKey   string     `gorm:"size:64" json:"-"`
	PresharedKey string     `gorm:"size:64" json:"-"`
	AllowedIPs   string     `gorm:"size:500" json:"allowed_ips"`
	Address      string     `gorm:"size:50" json:"address"`
	DNS          string     `gorm:"size:200" json:"dns"`
	Enabled      bool       `gorm:"default:true" json:"enabled"`
	ExpiresAt    *time.Time `json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// TableName returns the table name for Peer.
func (Peer) TableName() string {
	return "peers"
}
