package adapters

import (
	"fmt"

	"go.uber.org/zap"
)

// AdapterType represents the type of video generation adapter
type AdapterType string

const (
	AdapterTypeKling AdapterType = "kling"
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
	default:
		return nil, fmt.Errorf("unknown adapter type: %s", adapterType)
	}
}

// GetDefaultAdapter returns the default adapter (Kling v2.5 Turbo)
func (f *AdapterFactory) GetDefaultAdapter() VideoGeneratorAdapter {
	return NewKlingAdapter(f.replicateToken, f.logger)
}
