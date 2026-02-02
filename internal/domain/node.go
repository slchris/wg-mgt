package domain

import "time"

// NodeStatus represents the operational status of a WireGuard node.
type NodeStatus string

const (
	NodeStatusOnline  NodeStatus = "online"
	NodeStatusOffline NodeStatus = "offline"
	NodeStatusUnknown NodeStatus = "unknown"
)

// Node represents a WireGuard server node (VPS).
type Node struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Host        string     `gorm:"size:255;not null" json:"host"`
	SSHPort     int        `gorm:"default:22" json:"ssh_port"`
	SSHUser     string     `gorm:"size:100;not null" json:"ssh_user"`
	SSHKey      string     `gorm:"size:4096" json:"-"`
	NetworkID   *uint      `json:"network_id"`
	Network     *Network   `gorm:"foreignKey:NetworkID" json:"network,omitempty"`
	WGInterface string     `gorm:"size:20;default:'wg0'" json:"wg_interface"`
	WGPort      int        `gorm:"default:51820" json:"wg_port"`
	WGAddress   string     `gorm:"size:50" json:"wg_address"`
	PublicKey   string     `gorm:"size:64" json:"public_key"`
	PrivateKey  string     `gorm:"size:64" json:"-"`
	Endpoint    string     `gorm:"size:255" json:"endpoint"`
	Status      NodeStatus `gorm:"size:20;default:'unknown'" json:"status"`
	LastSeen    *time.Time `json:"last_seen"`
	Peers       []Peer     `gorm:"foreignKey:NodeID" json:"peers,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName returns the table name for Node.
func (Node) TableName() string {
	return "nodes"
}
