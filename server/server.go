package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dimas292/go-pkg/auth"
	"github.com/dimas292/go-pkg/config"
	"github.com/dimas292/go-pkg/database"
	"github.com/dimas292/go-pkg/logger"
	"github.com/dimas292/go-pkg/router"
	"github.com/dimas292/go-pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Server holds all shared dependencies and the Gin engine.
type Server struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client
	JWT    *auth.JWTService
	Router *gin.Engine
}

// New initializes the server: loads config, connects databases, sets up the router.
func New(configPath string) *Server {
	// Init logger first — everything else uses it
	logger.Init()

	// Init custom validators
	validator.Init()

	// Load config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	// Init Postgres
	db, err := database.InitPostgres(cfg.App.Db.Postgres)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect postgres")
	}
	logger.Info().Msg("postgres connected")

	// Init Redis
	rdb, err := database.InitRedis(cfg.App.Db.Redis)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect redis")
	}
	logger.Info().Msg("redis connected")

	// Init JWT
	jwtService := auth.NewJWTService(cfg.App.Jwt)
	logger.Info().Msg("jwt initialized")

	// Gin engine
	r := gin.Default()

	srv := &Server{
		Config: cfg,
		DB:     db,
		Redis:  rdb,
		JWT:    jwtService,
		Router: r,
	}

	// Register health check endpoint
	srv.registerHealthCheck()

	return srv
}

// registerHealthCheck registers the GET /health endpoint.
func (s *Server) registerHealthCheck() {
	s.Router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "healthy",
		})
	})
}

// RegisterModules registers feature modules under /api/v1.
func (s *Server) RegisterModules(modules ...router.Module) {
	router.RegisterModules(s.Router, "/api/v1", modules...)
}

// Run starts the HTTP server and waits for a shutdown signal (SIGINT/SIGTERM).
// On shutdown, it:
//  1. Stops accepting new connections
//  2. Waits for in-flight requests to finish (up to 10 seconds)
//  3. Closes PostgreSQL connection pool
//  4. Closes Redis connection
func (s *Server) Run() {
	port := s.Config.App.Port

	// Create an http.Server so we can call Shutdown() on it
	httpServer := &http.Server{
		Addr:    port,
		Handler: s.Router,
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.Info().Str("port", port).Msg("server started")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	// Wait for interrupt signal (CTRL+C or kill)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info().Str("signal", sig.String()).Msg("shutdown signal received")

	// Give in-flight requests up to 10 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Gracefully shut down HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("http server forced to shutdown")
	} else {
		logger.Info().Msg("http server stopped gracefully")
	}

	// 2. Close PostgreSQL connection pool
	if sqlDB, err := s.DB.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			logger.Error().Err(err).Msg("failed to close postgres")
		} else {
			logger.Info().Msg("postgres connection closed")
		}
	}

	// 3. Close Redis connection
	if err := s.Redis.Close(); err != nil {
		logger.Error().Err(err).Msg("failed to close redis")
	} else {
		logger.Info().Msg("redis connection closed")
	}

	logger.Info().Msg("server exited cleanly")
}

