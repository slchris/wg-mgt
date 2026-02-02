package repository

import (
	"github.com/slchris/wg-mgt/internal/domain"
	"gorm.io/gorm"
)

// UserRepository handles database operations for users.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user.
func (r *UserRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username.
func (r *UserRepository) GetByUsername(username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user.
func (r *UserRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

// Count returns the number of users.
func (r *UserRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&domain.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
