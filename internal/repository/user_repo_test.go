package repository

import (
	"testing"

	"github.com/slchris/wg-mgt/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &domain.User{
		Username:     "testuser",
		PasswordHash: "hash123",
		Role:         "admin",
	}

	err := repo.Create(user)
	require.NoError(t, err)
	assert.NotZero(t, user.ID)
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &domain.User{
		Username:     "testuser",
		PasswordHash: "hash123",
		Role:         "admin",
	}
	require.NoError(t, repo.Create(user))

	found, err := repo.GetByUsername("testuser")
	require.NoError(t, err)
	assert.Equal(t, user.Username, found.Username)
	assert.Equal(t, user.Role, found.Role)
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.GetByUsername("nonexistent")
	assert.Error(t, err)
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &domain.User{
		Username:     "testuser",
		PasswordHash: "hash123",
		Role:         "admin",
	}
	require.NoError(t, repo.Create(user))

	found, err := repo.GetByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Username, found.Username)
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &domain.User{
		Username:     "testuser",
		PasswordHash: "hash123",
		Role:         "admin",
	}
	require.NoError(t, repo.Create(user))

	user.Role = "user"
	err := repo.Update(user)
	require.NoError(t, err)

	found, err := repo.GetByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, "user", found.Role)
}

func TestUserRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	count, err := repo.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	user := &domain.User{
		Username:     "testuser",
		PasswordHash: "hash123",
		Role:         "admin",
	}
	require.NoError(t, repo.Create(user))

	count, err = repo.Count()
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
