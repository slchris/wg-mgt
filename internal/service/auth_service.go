package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/slchris/wg-mgt/internal/domain"
	"github.com/slchris/wg-mgt/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication operations.
type AuthService struct {
	userRepo   *repository.UserRepository
	jwtSecret  string
	expiration time.Duration
}

// Claims represents JWT claims.
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, expiration time.Duration) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtSecret:  jwtSecret,
		expiration: expiration,
	}
}

// Register creates a new user.
func (s *AuthService) Register(username, password, role string) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token.
func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	return s.generateToken(user)
}

// generateToken generates a JWT token for a user.
func (s *AuthService) generateToken(user *domain.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ValidateToken validates a JWT token and returns the claims.
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// IsFirstUser checks if no users exist yet.
func (s *AuthService) IsFirstUser() (bool, error) {
	count, err := s.userRepo.Count()
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// ChangePassword changes a user's password.
func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	return s.userRepo.Update(user)
}
