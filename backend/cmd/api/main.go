package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	_ "github.com/omnigen/backend/docs" // Import generated docs
	"github.com/omnigen/backend/internal/api"
	"github.com/omnigen/backend/internal/auth"
	"github.com/omnigen/backend/internal/aws"
	"github.com/omnigen/backend/internal/repository"
	"github.com/omnigen/backend/internal/service"
	"github.com/omnigen/backend/pkg/logger"
	"go.uber.org/zap"
)

// @title OmniGen API
// @version 1.0
// @description AI video generation pipeline API for creating professional-quality video content
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.omnigen.ai/support
// @contact.email support@omnigen.ai

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token authentication. Format: "Bearer {token}"

func main() {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	zapLogger, err := logger.NewLogger(cfg.Environment)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	zapLogger.Info("Starting OmniGen API",
		zap.String("environment", cfg.Environment),
		zap.String("port", cfg.Port),
		zap.String("aws_region", cfg.AWSRegion),
	)

	// Initialize AWS SDK configuration
	awsConfig, err := aws.NewConfig(cfg.AWSRegion)
	if err != nil {
		zapLogger.Fatal("Failed to initialize AWS config", zap.Error(err))
	}

	// Initialize AWS clients
	awsClients := aws.NewClients(awsConfig, cfg)
	zapLogger.Info("AWS clients initialized successfully")

	// Initialize repositories
	jobRepo := repository.NewDynamoDBRepository(
		awsClients.DynamoDB,
		cfg.JobTable,
		zapLogger,
	)

	s3Service := repository.NewS3Service(
		awsClients.S3,
		cfg.AssetsBucket,
		zapLogger,
	)

	usageRepo := repository.NewUsageRepository(
		awsClients.DynamoDB,
		cfg.UsageTable,
		zapLogger,
	)

	// Initialize rate limiter (1-minute window)
	rateLimiter := auth.NewRateLimiter(time.Minute, zapLogger)

	// Initialize services
	secretsService := service.NewSecretsService(
		awsClients.SecretsManager,
		cfg.ReplicateSecretARN,
		zapLogger,
	)

	// Retrieve API keys from Secrets Manager
	apiKeys, err := secretsService.GetAPIKeys()
	if err != nil {
		zapLogger.Fatal("Failed to retrieve API keys from Secrets Manager", zap.Error(err))
	}
	zapLogger.Info("API keys loaded successfully", zap.Int("count", len(apiKeys)))

	stepFunctionsService := service.NewStepFunctionsService(
		awsClients.StepFunctions,
		cfg.StepFunctionsARN,
		zapLogger,
	)

	promptParser := service.NewPromptParser(zapLogger)
	scenePlanner := service.NewScenePlanner(zapLogger)

	generatorService := service.NewGeneratorService(
		jobRepo,
		stepFunctionsService,
		promptParser,
		scenePlanner,
		zapLogger,
	)

	// Initialize JWT validator
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json",
		cfg.AWSRegion, cfg.CognitoUserPoolID)

	jwtValidator := auth.NewJWTValidator(jwksURL, cfg.JWTIssuer, cfg.CognitoClientID, zapLogger)

	// Fetch JWKS keys at startup
	if err := jwtValidator.FetchJWKS(); err != nil {
		zapLogger.Fatal("Failed to fetch JWKS", zap.Error(err))
	}
	zapLogger.Info("JWT validator initialized successfully")

	// Cookie configuration for httpOnly tokens
	cookieConfig := auth.CookieConfig{
		Secure:   cfg.Environment == "production", // HTTPS only in production
		Domain:   "",                               // Empty for same-origin cookies
		SameSite: http.SameSiteStrictMode,         // Strict CSRF protection
	}

	// Initialize HTTP server
	server := api.NewServer(&api.ServerConfig{
		Port:             cfg.Port,
		Environment:      cfg.Environment,
		Logger:           zapLogger,
		JobRepo:          jobRepo,
		S3Service:        s3Service,
		UsageRepo:        usageRepo,
		GeneratorService: generatorService,
		APIKeys:          apiKeys,
		JWTValidator:     jwtValidator,
		RateLimiter:      rateLimiter,
		CookieConfig:     cookieConfig,
		CloudFrontDomain: cfg.CloudFrontDomain,
		CognitoDomain:    cfg.CognitoDomain,
		ReadTimeout:      time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:     time.Duration(cfg.WriteTimeout) * time.Second,
	})

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      server.Router(),
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		zapLogger.Info("Starting HTTP server", zap.String("address", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("Server exited cleanly")
}

// Config holds all application configuration
type Config struct {
	// Server configuration
	Port         string `envconfig:"PORT" default:"8080"`
	Environment  string `envconfig:"ENVIRONMENT" default:"production"`
	ReadTimeout  int    `envconfig:"READ_TIMEOUT" default:"30"`
	WriteTimeout int    `envconfig:"WRITE_TIMEOUT" default:"30"`

	// AWS configuration
	AWSRegion          string `envconfig:"AWS_REGION" required:"true"`
	AssetsBucket       string `envconfig:"ASSETS_BUCKET" required:"true"`
	JobTable           string `envconfig:"JOB_TABLE" required:"true"`
	UsageTable         string `envconfig:"USAGE_TABLE" required:"true"`
	StepFunctionsARN   string `envconfig:"STEP_FUNCTIONS_ARN" required:"true"`
	ReplicateSecretARN string `envconfig:"REPLICATE_SECRET_ARN" required:"true"`

	// Authentication configuration
	CognitoUserPoolID string `envconfig:"COGNITO_USER_POOL_ID" required:"true"`
	CognitoClientID   string `envconfig:"COGNITO_CLIENT_ID" required:"true"`
	JWTIssuer         string `envconfig:"JWT_ISSUER" required:"true"`
	CognitoDomain     string `envconfig:"COGNITO_DOMAIN"` // Optional: for CORS

	// Frontend configuration (optional, for CORS)
	CloudFrontDomain string `envconfig:"CLOUDFRONT_DOMAIN"`
}

func loadConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %w", err)
	}
	return &cfg, nil
}
