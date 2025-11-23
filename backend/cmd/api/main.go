package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/omnigen/backend/docs" // Import generated docs
	"github.com/omnigen/backend/internal/adapters"
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
// @description AI ad creative generation pipeline API for creating professional-quality ads
// @termsOfService http://swagger.io/terms/

// @BasePath /api/v1

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

	// Check system dependencies
	if err := checkDependencies(); err != nil {
		zapLogger.Fatal("System dependency check failed", zap.Error(err))
	}
	zapLogger.Info("System dependencies verified (ffmpeg found)")

	// Initialize AWS SDK configuration
	awsConfig, err := aws.NewConfig(context.Background(), cfg.AWSRegion)
	if err != nil {
		zapLogger.Fatal("Failed to initialize AWS config", zap.Error(err))
	}

	// Initialize AWS clients
	awsClients := aws.NewClients(awsConfig)
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

	brandGuidelinesRepo := repository.NewDynamoDBBrandGuidelinesRepository(
		awsClients.DynamoDB,
		cfg.BrandGuidelinesTable,
		zapLogger,
	)

	// Initialize services
	secretsService := service.NewSecretsService(
		awsClients.SecretsManager,
		cfg.ReplicateSecretARN,
		cfg.OpenAISecretARN,
		zapLogger,
	)

	// Retrieve API keys from Secrets Manager
	apiKeys, err := secretsService.GetAPIKeys(context.Background())
	if err != nil {
		zapLogger.Fatal("Failed to retrieve API keys from Secrets Manager", zap.Error(err))
	}
	zapLogger.Info("API keys loaded successfully", zap.Int("count", len(apiKeys)))

	// Initialize parser service for script generation with GPT-4o
	// Get the Replicate API key from Secrets Manager
	replicateAPIKey, err := secretsService.GetReplicateAPIKey(context.Background())
	if err != nil {
		zapLogger.Fatal("Failed to retrieve Replicate API key", zap.Error(err))
	}

	// Create GPT-4o adapter for intelligent script generation
	gpt4oAdapter := adapters.NewGPT4oAdapter(replicateAPIKey, zapLogger)

	parserService := service.NewParserService(
		gpt4oAdapter,
		zapLogger,
	)
	zapLogger.Info("Parser service initialized with GPT-4o")

	// Initialize Asset Service
	assetService := service.NewAssetService(
		s3Service,
		zapLogger,
	)
	zapLogger.Info("Asset service initialized")

	// Initialize video and audio generation adapters
	veoAdapter := adapters.NewVeoAdapter(replicateAPIKey, zapLogger)
	minimaxAdapter := adapters.NewMinimaxAdapter(replicateAPIKey, zapLogger)
	zapLogger.Info("Video and audio generation adapters initialized (Veo 3.1)")

	// Initialize TTS adapter for narrator voiceover generation
	// Try to get OpenAI API key from Secrets Manager or environment variable
	var ttsAdapter adapters.TTSAdapter
	openaiAPIKey, err := secretsService.GetOpenAIAPIKey(context.Background())
	if err != nil {
		zapLogger.Warn("OpenAI API key not available - narrator voiceover generation will not be available",
			zap.Error(err),
		)
		// ttsAdapter will remain nil - this is handled gracefully in generateNarratorVoiceover
	} else if openaiAPIKey != "" {
		ttsAdapter = adapters.NewOpenAITTSAdapter(openaiAPIKey, zapLogger)
		zapLogger.Info("TTS adapter initialized with OpenAI TTS API")
	} else {
		zapLogger.Warn("OPENAI_API_KEY not configured - narrator voiceover generation will not be available")
	}

	// Initialize JWT validator
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json",
		cfg.AWSRegion, cfg.CognitoUserPoolID)

	jwtValidator := auth.NewJWTValidator(jwksURL, cfg.JWTIssuer, cfg.CognitoClientID, zapLogger)

	// Fetch JWKS keys at startup
	if err := jwtValidator.FetchJWKS(); err != nil {
		if cfg.Environment == "production" {
			zapLogger.Fatal("Failed to fetch JWKS", zap.Error(err))
		} else {
			zapLogger.Warn("Failed to fetch JWKS (continuing in development mode)", zap.Error(err))
		}
	} else {
		zapLogger.Info("JWT validator initialized successfully")
	}

	// Cookie configuration for httpOnly tokens
	cookieConfig := auth.CookieConfig{
		Secure:   cfg.Environment == "production", // HTTPS only in production
		Domain:   "",                              // Empty for same-origin cookies
		SameSite: http.SameSiteLaxMode,            // Lax mode for production compatibility (allows top-level navigation)
	}

	// Initialize HTTP server with goroutine-based async architecture
	server := api.NewServer(&api.ServerConfig{
		Port:                cfg.Port,
		Environment:         cfg.Environment,
		Logger:              zapLogger,
		JobRepo:             jobRepo,
		S3Service:           s3Service,
		UsageRepo:           usageRepo,
		BrandGuidelinesRepo: brandGuidelinesRepo,
		ParserService:       parserService,
		AssetService:        assetService,
		VeoAdapter:          veoAdapter,     // Video generation (Veo 3.1)
		MinimaxAdapter:      minimaxAdapter, // Audio generation
		TTSAdapter:          ttsAdapter,     // Text-to-speech for narrator voiceover
		AssetsBucket:        cfg.AssetsBucket,
		APIKeys:             apiKeys,
		JWTValidator:        jwtValidator,
		CookieConfig:        cookieConfig,
		CloudFrontDomain:    cfg.CloudFrontDomain,
		CognitoDomain:       cfg.CognitoDomain,
		OpenAIKey:           cfg.OpenAIKey,
		ReadTimeout:         time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:        time.Duration(cfg.WriteTimeout) * time.Second,
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
	AWSRegion               string `envconfig:"AWS_REGION" required:"true"`
	AssetsBucket            string `envconfig:"ASSETS_BUCKET" required:"true"`
	JobTable                string `envconfig:"JOB_TABLE" required:"true"`
	UsageTable              string `envconfig:"USAGE_TABLE" required:"true"`
	BrandGuidelinesTable    string `envconfig:"BRAND_GUIDELINES_TABLE" default:"omnigen-brand-guidelines"`
	ReplicateSecretARN      string `envconfig:"REPLICATE_SECRET_ARN"` // Optional: if not set, will use REPLICATE_API_KEY env var
	OpenAISecretARN         string `envconfig:"OPENAI_SECRET_ARN"`    // Optional: if not set, will use OPENAI_API_KEY env var

	// Authentication configuration
	CognitoUserPoolID string `envconfig:"COGNITO_USER_POOL_ID" required:"true"`
	CognitoClientID   string `envconfig:"COGNITO_CLIENT_ID" required:"true"`
	JWTIssuer         string `envconfig:"JWT_ISSUER" required:"true"`
	CognitoDomain     string `envconfig:"COGNITO_DOMAIN"` // Optional: for CORS

	// Frontend configuration (optional, for CORS)
	CloudFrontDomain string `envconfig:"CLOUDFRONT_DOMAIN"`

	// OpenAI configuration (optional, for title generation)
	OpenAIKey string `envconfig:"GPT4O_API_KEY"` // OpenAI API key for title generation

	// TTS configuration (for narrator voiceover generation)
	TTSAPIKey string `envconfig:"TTS_API_KEY"` // OpenAI TTS API key for narrator voiceover
}

func loadConfig() (*Config, error) {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: Could not get working directory: %v", err)
		wd = "."
	}

	// Try to load .env.local first, then .env (for backwards compatibility)
	// Try multiple paths to handle different working directories
	envPaths := []string{
		".env.local",
		".env",
		filepath.Join(wd, ".env.local"),
		filepath.Join(wd, ".env"),
		filepath.Join(wd, "backend", ".env.local"),
		filepath.Join(wd, "backend", ".env"),
	}

	loaded := false
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			loaded = true
			log.Printf("Loaded environment variables from %s (working dir: %s)", path, wd)
			break
		}
	}

	if !loaded {
		log.Printf("No .env file found in working directory %s, using environment variables only", wd)
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process environment variables: %w", err)
	}
	return &cfg, nil
}

// checkDependencies verifies required system dependencies are installed
func checkDependencies() error {
	// Check for ffmpeg (required for video composition)
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found in PATH - required for video composition. Install with: apt-get install ffmpeg (Debian/Ubuntu) or apk add ffmpeg (Alpine)")
	}
	return nil
}
