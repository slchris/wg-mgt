package service

import (
	"testing"
	"time"

	"github.com/slchris/wg-mgt/internal/domain"
	"github.com/slchris/wg-mgt/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	return db
}

func TestAuthService_Register(t *testing.T) {
	db := setupAuthTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret", time.Hour)

	user, err := authService.Register("testuser", "password123", "admin")
	require.NoError(t, err)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "admin", user.Role)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, "password123", user.PasswordHash)
}

func TestAuthService_Register_DuplicateUsername(t *testing.T) {
	db := setupAuthTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret", time.Hour)

	_, err := authService.Register("testuser", "password123", "admin")
	require.NoError(t, err)

	_, err = authService.Register("testuser", "password456", "user")
	assert.Error(t, err)
}

func TestAuthService_Login(t *testing.T) {
	db := setupAuthTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret", time.Hour)

	_, err := authService.Register("testuser", "password123", "admin")
	require.NoError(t, err)

	token, err := authService.Login("testuser", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	db := setupAuthTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret", time.Hour)

	_, err := authService.Register("testuser", "password123", "admin")
	require.NoError(t, err)

	_, err = authService.Login("testuser", "wrongpassword")
	assert.Error(t, err)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	db := setupAuthTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret", time.Hour)

	_, err := authService.Login("nonexistent", "password123")
	assert.Error(t, err)
}

func TestAuthService_ValidateToken(t *testing.T) {
	db := setupAuthTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret", time.Hour)

	user, err := authService.Register("testuser", "password123", "admin")
	require.NoError(t, err)

	token, err := authService.Login("testuser", "password123")
	require.NoError(t, err)

	claims, err := authService.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "admin", claims.Role)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	db := setupAuthTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret", time.Hour)

	_, err := authService.ValidateToken("invalid-token")
	assert.Error(t, err)
}

func TestAuthService_ValidateToken_ExpiredToken(t *testing.T) {
	db := setupAuthTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret", -time.Hour)

	_, err := authService.Register("testuser", "password123", "admin")
	require.NoError(t, err)

	token, err := authService.Login("testuser", "password123")
	require.NoError(t, err)

	_, err = authService.ValidateToken(token)
	assert.Error(t, err)
}

func TestAuthService_IsFirstUser(t *testing.T) {
	db := setupAuthTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret", time.Hour)

	isFirst, err := authService.IsFirstUser()
	require.NoError(t, err)
	assert.True(t, isFirst)

	_, err = authService.Register("testuser", "password123", "admin")
	require.NoError(t, err)

	isFirst, err = authService.IsFirstUser()
	require.NoError(t, err)
	assert.False(t, isFirst)
}

func TestAuthService_ChangePassword(t *testing.T) {
	db := setupAuthTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret", time.Hour)

	user, err := authService.Register("testuser", "password123", "admin")
	require.NoError(t, err)

	err = authService.ChangePassword(user.ID, "password123", "newpassword456")
	require.NoError(t, err)

	_, err = authService.Login("testuser", "newpassword456")
	require.NoError(t, err)

	_, err = authService.Login("testuser", "password123")
	assert.Error(t, err)
}

func TestAuthService_ChangePassword_WrongOldPassword(t *testing.T) {
	db := setupAuthTestDB(t)
	userRepo := repository.NewUserRepository(db)
	authService := NewAuthService(userRepo, "test-secret", time.Hour)

	user, err := authService.Register("testuser", "password123", "admin")
	require.NoError(t, err)

	err = authService.ChangePassword(user.ID, "wrongpassword", "newpassword456")
	assert.Error(t, err)
}
