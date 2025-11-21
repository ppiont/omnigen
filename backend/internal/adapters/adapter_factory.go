package adapters

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

// AdapterType represents the type of video generation adapter
type AdapterType string

const (
	AdapterTypeKling AdapterType = "kling"
	AdapterTypeVeo   AdapterType = "veo"
	// Future adapters:
	// AdapterTypeRunway AdapterType = "runway"
	// AdapterTypePika AdapterType = "pika"
)

// AdapterFactory creates video generation adapters
type AdapterFactory struct {
	replicateToken string
	logger         *zap.Logger
}

// NewAdapterFactory creates a new adapter factory
func NewAdapterFactory(replicateToken string, logger *zap.Logger) *AdapterFactory {
	return &AdapterFactory{
		replicateToken: replicateToken,
		logger:         logger,
	}
}

// CreateAdapter creates a video generation adapter of the specified type
func (f *AdapterFactory) CreateAdapter(adapterType AdapterType) (VideoGeneratorAdapter, error) {
	switch adapterType {
	case AdapterTypeKling:
		return NewKlingAdapter(f.replicateToken, f.logger), nil
	case AdapterTypeVeo:
		return NewVeoAdapter(f.replicateToken, f.logger), nil
	default:
		return nil, fmt.Errorf("unknown adapter type: %s", adapterType)
	}
}

// GetDefaultAdapter returns the adapter based on environment configuration
func (f *AdapterFactory) GetDefaultAdapter() VideoGeneratorAdapter {
	// Check environment variable for adapter type
	adapterType := os.Getenv("VIDEO_ADAPTER_TYPE")
	switch adapterType {
	case "kling":
		f.logger.Info("Using Kling adapter based on environment configuration")
		return NewKlingAdapter(f.replicateToken, f.logger)
	case "veo", "":
		f.logger.Info("Using Veo 3.1 adapter (default)")
		return NewVeoAdapter(f.replicateToken, f.logger)
	default:
		f.logger.Warn("Unknown VIDEO_ADAPTER_TYPE, falling back to Veo",
			zap.String("adapter_type", adapterType))
		return NewVeoAdapter(f.replicateToken, f.logger)
	}
}
