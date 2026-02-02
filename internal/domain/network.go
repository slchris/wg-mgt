package domain

import "time"

// Network represents a virtual network configuration.
type Network struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null;uniqueIndex" json:"name"`
	CIDR        string    `gorm:"size:50;not null" json:"cidr"`
	Gateway     string    `gorm:"size:50" json:"gateway"`
	Description string    `gorm:"size:500" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName returns the table name for Network.
func (Network) TableName() string {
	return "networks"
}
