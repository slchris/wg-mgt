package repository

import (
	"testing"

	"github.com/slchris/wg-mgt/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&domain.Node{}, &domain.Peer{}, &domain.Network{}, &domain.User{})
	require.NoError(t, err)

	return db
}

func TestNodeRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNodeRepository(db)

	node := &domain.Node{
		Name:       "test-node",
		Host:       "192.168.1.1",
		PublicKey:  "test-public-key",
		PrivateKey: "test-private-key",
		SSHUser:    "root",
	}

	err := repo.Create(node)
	require.NoError(t, err)
	assert.NotZero(t, node.ID)
}

func TestNodeRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNodeRepository(db)

	node := &domain.Node{
		Name:       "test-node",
		Host:       "192.168.1.1",
		PublicKey:  "test-public-key",
		PrivateKey: "test-private-key",
		SSHUser:    "root",
	}
	require.NoError(t, repo.Create(node))

	found, err := repo.GetByID(node.ID)
	require.NoError(t, err)
	assert.Equal(t, node.Name, found.Name)
	assert.Equal(t, node.Host, found.Host)
}

func TestNodeRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNodeRepository(db)

	_, err := repo.GetByID(9999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestNodeRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNodeRepository(db)

	nodes := []*domain.Node{
		{Name: "node1", Host: "192.168.1.1", PublicKey: "key1", PrivateKey: "priv1", SSHUser: "root"},
		{Name: "node2", Host: "192.168.1.2", PublicKey: "key2", PrivateKey: "priv2", SSHUser: "root"},
	}
	for _, n := range nodes {
		require.NoError(t, repo.Create(n))
	}

	result, err := repo.GetAll()
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestNodeRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNodeRepository(db)

	node := &domain.Node{
		Name:       "test-node",
		Host:       "192.168.1.1",
		PublicKey:  "test-public-key",
		PrivateKey: "test-private-key",
		SSHUser:    "root",
	}
	require.NoError(t, repo.Create(node))

	node.Name = "updated-node"
	err := repo.Update(node)
	require.NoError(t, err)

	found, err := repo.GetByID(node.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated-node", found.Name)
}

func TestNodeRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewNodeRepository(db)

	node := &domain.Node{
		Name:       "test-node",
		Host:       "192.168.1.1",
		PublicKey:  "test-public-key",
		PrivateKey: "test-private-key",
		SSHUser:    "root",
	}
	require.NoError(t, repo.Create(node))

	err := repo.Delete(node.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(node.ID)
	assert.Error(t, err)
}
