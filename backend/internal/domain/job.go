package domain

// Job represents a video generation job
type Job struct {
	JobID        string  `dynamodbav:"job_id" json:"job_id"`
	UserID       string  `dynamodbav:"user_id" json:"user_id"`
	Status       string  `dynamodbav:"status" json:"status"` // pending, processing, completed, failed
	Prompt       string  `dynamodbav:"prompt" json:"prompt"`
	Duration     int     `dynamodbav:"duration" json:"duration"`
	Style        string  `dynamodbav:"style" json:"style"`
	AspectRatio  string  `dynamodbav:"aspect_ratio" json:"aspect_ratio"`
	VideoKey     string  `dynamodbav:"video_key" json:"video_key"` // S3 key
	CreatedAt    int64   `dynamodbav:"created_at" json:"created_at"`
	CompletedAt  *int64  `dynamodbav:"completed_at,omitempty" json:"completed_at,omitempty"`
	ErrorMessage *string `dynamodbav:"error_message,omitempty" json:"error_message,omitempty"`
	TTL          int64   `dynamodbav:"ttl" json:"ttl"` // Unix timestamp for auto-deletion
}

// Scene represents a single scene in the video
type Scene struct {
	Number     int     `json:"number"`
	Duration   float64 `json:"duration"` // seconds
	Prompt     string  `json:"prompt"`
	Style      string  `json:"style"`
	Transition string  `json:"transition"` // fade, cut, dissolve
}

// GenerateRequest represents a video generation request
type GenerateRequest struct {
	UserID      string
	Prompt      string
	Duration    int
	AspectRatio string
	Style       string
}

// ParsedPrompt represents the parsed components of a prompt
type ParsedPrompt struct {
	ProductType  string   `json:"product_type"`
	VisualStyle  []string `json:"visual_style"`
	ColorPalette []string `json:"color_palette"`
	AspectRatio  string   `json:"aspect_ratio"`
	Duration     int      `json:"duration"`
	TextOverlays []string `json:"text_overlays"`
}

// StepFunctionsInput represents the input to Step Functions workflow
type StepFunctionsInput struct {
	JobID    string   `json:"job_id"`
	Prompt   string   `json:"prompt"`
	Duration int      `json:"duration"`
	Style    string   `json:"style"`
	Scenes   []Scene  `json:"scenes"`
}

// JobStatus constants
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)

// AspectRatio constants
const (
	AspectRatio16x9 = "16:9"
	AspectRatio9x16 = "9:16"
	AspectRatio1x1  = "1:1"
)
