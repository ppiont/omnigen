package domain

// Job represents a video generation job
type Job struct {
	JobID        string  `dynamodbav:"job_id" json:"job_id"`
	UserID       string  `dynamodbav:"user_id" json:"user_id"`
	ScriptID     string  `dynamodbav:"script_id,omitempty" json:"script_id,omitempty"` // Optional script ID (for new workflow)
	Status       string  `dynamodbav:"status" json:"status"`                           // pending, processing, completed, failed
	Prompt       string  `dynamodbav:"prompt,omitempty" json:"prompt,omitempty"`       // Legacy field
	Duration     int     `dynamodbav:"duration,omitempty" json:"duration,omitempty"`
	Style        string  `dynamodbav:"style,omitempty" json:"style,omitempty"`
	AspectRatio  string  `dynamodbav:"aspect_ratio,omitempty" json:"aspect_ratio,omitempty"`
	VideoKey     string  `dynamodbav:"video_key,omitempty" json:"video_key,omitempty"` // S3 key
	CreatedAt    int64   `dynamodbav:"created_at" json:"created_at"`
	CompletedAt  *int64  `dynamodbav:"completed_at,omitempty" json:"completed_at,omitempty"`
	ErrorMessage *string `dynamodbav:"error_message,omitempty" json:"error_message,omitempty"`
	TTL          int64   `dynamodbav:"ttl" json:"ttl"` // Unix timestamp for auto-deletion
}

// GenerateRequest represents a video generation request
type GenerateRequest struct {
	UserID        string
	Prompt        string
	Duration      int
	AspectRatio   string
	Style         string
	StartImageURL string
	EnableAudio   bool // Whether to generate audio for this video
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

// StepFunctionsInput represents the simplified input to Step Functions workflow
type StepFunctionsInput struct {
	JobID       string `json:"job_id"`
	Prompt      string `json:"prompt"`
	Duration    int    `json:"duration"`     // Total duration in seconds
	AspectRatio string `json:"aspect_ratio"` // "16:9", "9:16", or "1:1"
	StartImage  string `json:"start_image,omitempty"`
	NumClips    int    `json:"num_clips"`   // Calculated: duration / 10
	MusicMood   string `json:"music_mood"`  // upbeat, calm, dramatic, energetic
	MusicStyle  string `json:"music_style"` // electronic, acoustic, orchestral
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
