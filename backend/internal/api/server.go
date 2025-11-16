package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/api/handlers"
	"github.com/omnigen/backend/internal/api/middleware"
	"github.com/omnigen/backend/internal/auth"
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
	MockMode         bool // Enable mock responses for frontend development
	Logger           *zap.Logger
	JobRepo          *repository.DynamoDBRepository
	S3Service        *repository.S3Service
	UsageRepo        *repository.UsageRepository
	GeneratorService *service.GeneratorService
	ParserService    *service.ParserService // Script generation service
	MockService      *service.MockService   // Mock service for frontend development
	APIKeys          []string               // Deprecated: Use JWTValidator instead
	JWTValidator     *auth.JWTValidator
	RateLimiter      *auth.RateLimiter
	CookieConfig     auth.CookieConfig // Cookie configuration for httpOnly tokens
	CloudFrontDomain string            // For CORS in production
	CognitoDomain    string            // Cognito hosted UI domain for CORS
	LambdaClient     interface{}       // Lambda client for async script generation
	LambdaParserARN  string            // Parser Lambda ARN
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
	// Build allowed origins list
	allowedOrigins := []string{
		"http://localhost:3000", // Local development
		"http://localhost:5173", // Vite default port
		"http://localhost:8080", // Local backend
	}

	// Add CloudFront domain if in production
	if config.Environment == "production" && config.CloudFrontDomain != "" {
		allowedOrigins = append(allowedOrigins, "https://"+config.CloudFrontDomain)
	}

	// Add Cognito hosted UI domain for OAuth2 redirects
	if config.CognitoDomain != "" {
		allowedOrigins = append(allowedOrigins, config.CognitoDomain)
	}

	corsConfig := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
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
	s.router.HEAD("/health", healthHandler.Check) // For Docker HEALTHCHECK

	// Swagger documentation (no auth required for development)
	if s.config.Environment != "production" {
		s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Auth routes (no JWT middleware - used for login/logout)
	authHandler := handlers.NewAuthHandler(
		s.config.JWTValidator,
		s.config.CookieConfig,
		s.config.Logger,
	)
	authGroup := s.router.Group("/api/v1/auth")
	{
		authGroup.POST("/login", authHandler.Login)                                                          // Exchange Cognito tokens for cookies
		authGroup.POST("/refresh", authHandler.Refresh)                                                      // Refresh token endpoint
		authGroup.POST("/logout", authHandler.Logout)                                                        // Clear cookies
		authGroup.GET("/me", auth.JWTAuthMiddleware(s.config.JWTValidator, s.config.Logger), authHandler.Me) // Get current user (requires auth)
	}

	// API v1 routes
	v1 := s.router.Group("/api/v1")

	// Only apply auth and rate limiting in non-mock mode
	if !s.config.MockMode {
		v1.Use(auth.JWTAuthMiddleware(s.config.JWTValidator, s.config.Logger))
		v1.Use(auth.RateLimitMiddleware(s.config.RateLimiter, s.config.Logger))
		s.config.Logger.Info("Auth and rate limiting enabled for API routes")
	} else {
		s.config.Logger.Warn("MOCK MODE: Auth and rate limiting disabled - API is publicly accessible")
	}

	{
		// Initialize handlers
		generateHandler := handlers.NewGenerateHandler(
			s.config.GeneratorService,
			s.config.MockService,
			s.config.Logger,
			s.config.MockMode,
		)

		jobsHandler := handlers.NewJobsHandler(
			s.config.JobRepo,
			s.config.S3Service,
			s.config.MockService,
			s.config.Logger,
			s.config.MockMode,
		)

		progressHandler := handlers.NewProgressHandler(
			s.config.MockService,
			s.config.Logger,
			s.config.MockMode,
		)

		presetsHandler := handlers.NewPresetsHandler(
			s.config.MockService,
			s.config.Logger,
		)

		// Type assert Lambda client
		var lambdaClient interface{}
		if s.config.LambdaClient != nil {
			lambdaClient = s.config.LambdaClient
		}

		parserHandler := handlers.NewParserHandler(
			s.config.ParserService,
			lambdaClient,
			s.config.LambdaParserARN,
			s.config.Logger,
			s.config.MockMode,
		)

		// Routes with quota enforcement for generation endpoint (skipped in mock mode)
		if !s.config.MockMode {
			v1.POST("/generate",
				auth.QuotaEnforcementMiddleware(s.config.UsageRepo, s.config.Logger),
				generateHandler.Generate,
			)
			v1.POST("/parse",
				auth.QuotaEnforcementMiddleware(s.config.UsageRepo, s.config.Logger),
				parserHandler.Parse,
			)
		} else {
			v1.POST("/generate", generateHandler.Generate)
			v1.POST("/parse", parserHandler.Parse)
		}

		// Script routes (GET/PUT - no quota enforcement needed)
		v1.GET("/scripts/:id", parserHandler.GetScript)
		v1.PUT("/scripts/:id", parserHandler.UpdateScript)

		// Job routes
		v1.GET("/jobs/:id", jobsHandler.GetJob)
		v1.GET("/jobs", jobsHandler.ListJobs)
		v1.GET("/jobs/:id/progress", progressHandler.GetProgress)
		v1.GET("/presets", presetsHandler.GetPresets)
	}
}
