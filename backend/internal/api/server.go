package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/api/handlers"
	"github.com/omnigen/backend/internal/api/middleware"
	"github.com/omnigen/backend/internal/repository"
	"github.com/omnigen/backend/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// ServerConfig holds the server configuration
type ServerConfig struct {
	Port             string
	Environment      string
	Logger           *zap.Logger
	JobRepo          *repository.DynamoDBRepository
	S3Service        *repository.S3Service
	GeneratorService *service.GeneratorService
	APIKeys          []string
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
}

// Server represents the HTTP server
type Server struct {
	config *ServerConfig
	router *gin.Engine
}

// NewServer creates a new HTTP server
func NewServer(config *ServerConfig) *Server {
	// Set Gin mode based on environment
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	// Add middlewares
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(config.Logger))

	// CORS configuration
	corsConfig := cors.Config{
		AllowOrigins:     []string{"*"}, // Configure this for production
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "x-api-key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	s := &Server{
		config: config,
		router: router,
	}

	// Setup routes
	s.setupRoutes()

	return s
}

// Router returns the Gin router
func (s *Server) Router() *gin.Engine {
	return s.router
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Health check endpoint (no auth required)
	healthHandler := handlers.NewHealthHandler(
		s.config.JobRepo,
		s.config.S3Service,
		s.config.Logger,
	)
	s.router.GET("/health", healthHandler.Check)

	// Swagger documentation (no auth required for development)
	if s.config.Environment != "production" {
		s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// API v1 routes (with authentication)
	v1 := s.router.Group("/api/v1")
	v1.Use(middleware.Auth(s.config.APIKeys))
	{
		// Initialize handlers
		generateHandler := handlers.NewGenerateHandler(
			s.config.GeneratorService,
			s.config.Logger,
		)

		jobsHandler := handlers.NewJobsHandler(
			s.config.JobRepo,
			s.config.S3Service,
			s.config.Logger,
		)

		// Routes
		v1.POST("/generate", generateHandler.Generate)
		v1.GET("/jobs/:id", jobsHandler.GetJob)
		v1.GET("/jobs", jobsHandler.ListJobs)
	}
}
