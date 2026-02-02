package app

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/slchris/wg-mgt/internal/config"
	"github.com/slchris/wg-mgt/internal/domain"
	"github.com/slchris/wg-mgt/internal/handler"
	"github.com/slchris/wg-mgt/internal/repository"
	"github.com/slchris/wg-mgt/internal/router"
	"github.com/slchris/wg-mgt/internal/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// App represents the application.
type App struct {
	config *config.Config
	db     *gorm.DB
	router http.Handler
}

// New creates a new application.
func New(webFS fs.FS) (*App, error) {
	cfg := config.Load()

	db, err := initDB(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Repositories
	nodeRepo := repository.NewNodeRepository(db)
	peerRepo := repository.NewPeerRepository(db)
	networkRepo := repository.NewNetworkRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Services
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.Expiration)
	nodeService := service.NewNodeService(nodeRepo)
	peerService := service.NewPeerService(peerRepo, nodeRepo)
	networkService := service.NewNetworkService(networkRepo)

	// Auto-create admin user if configured and no users exist
	if cfg.Admin.Username != "" && cfg.Admin.Password != "" {
		isFirst, err := authService.IsFirstUser()
		if err != nil {
			return nil, fmt.Errorf("failed to check for existing users: %w", err)
		}
		if isFirst {
			_, err := authService.Register(cfg.Admin.Username, cfg.Admin.Password, "admin")
			if err != nil {
				return nil, fmt.Errorf("failed to create admin user: %w", err)
			}
			log.Printf("Created default admin user: %s", cfg.Admin.Username)
		}
	}

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	nodeHandler := handler.NewNodeHandler(nodeService)
	peerHandler := handler.NewPeerHandler(peerService)
	networkHandler := handler.NewNetworkHandler(networkService)

	// Router
	r := router.New(&router.Config{
		LocalOnly:      cfg.Server.LocalOnly,
		ServerPort:     cfg.Server.Port,
		ServerHost:     cfg.Server.Host,
		AuthService:    authService,
		NodeHandler:    nodeHandler,
		PeerHandler:    peerHandler,
		NetworkHandler: networkHandler,
		AuthHandler:    authHandler,
		WebFS:          webFS,
	})

	return &App{
		config: cfg,
		db:     db,
		router: r,
	}, nil
}

// Run starts the application.
func (a *App) Run() error {
	addr := fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port)
	log.Printf("Starting server on %s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      a.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server.ListenAndServe()
}

// initDB initializes the database.
func initDB(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// Auto migrate
	if err := db.AutoMigrate(
		&domain.Network{},
		&domain.Node{},
		&domain.Peer{},
		&domain.User{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
