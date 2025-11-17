package domain

// Job represents a video generation job
type Job struct {
	// Identity
	JobID  string `dynamodbav:"job_id" json:"job_id"`
	UserID string `dynamodbav:"user_id" json:"user_id"`

	// Status tracking
	Status string `dynamodbav:"status" json:"status"`                   // pending, processing, completed, failed
	Stage  string `dynamodbav:"stage,omitempty" json:"stage,omitempty"` // Granular progress: script_generating, scene_1_complete, etc.

	// Dynamic progress data (structured progress tracking)
	Metadata map[string]interface{} `dynamodbav:"metadata,omitempty" json:"metadata,omitempty"` // Stores thumbnails, progress data, etc.

	// Progress fields (NEW - structured for better API responses)
	ThumbnailURL    string   `dynamodbav:"thumbnail_url,omitempty" json:"thumbnail_url,omitempty"`       // Preview thumbnail from first scene
	AudioURL        string   `dynamodbav:"audio_url,omitempty" json:"audio_url,omitempty"`               // Generated audio track URL
	ScenesCompleted int      `dynamodbav:"scenes_completed,omitempty" json:"scenes_completed,omitempty"` // Number of completed scenes
	SceneVideoURLs  []string `dynamodbav:"scene_video_urls,omitempty" json:"scene_video_urls,omitempty"` // Individual scene video URLs

	// Input parameters
	Prompt      string `dynamodbav:"prompt,omitempty" json:"prompt,omitempty"`
	Duration    int    `dynamodbav:"duration,omitempty" json:"duration,omitempty"`
	AspectRatio string `dynamodbav:"aspect_ratio,omitempty" json:"aspect_ratio,omitempty"`

	// Embedded script (replaces separate scripts table)
	Title          string    `dynamodbav:"title,omitempty" json:"title,omitempty"`
	Scenes         []Scene   `dynamodbav:"scenes,omitempty" json:"scenes,omitempty"`
	AudioSpec      AudioSpec `dynamodbav:"audio_spec,omitempty" json:"audio_spec,omitempty"`
	ScriptMetadata Metadata  `dynamodbav:"script_metadata,omitempty" json:"script_metadata,omitempty"`

	// Output
	VideoKey     string  `dynamodbav:"video_key,omitempty" json:"video_key,omitempty"` // S3 key
	ErrorMessage *string `dynamodbav:"error_message,omitempty" json:"error_message,omitempty"`

	// Timestamps (all Unix timestamps)
	CreatedAt   int64  `dynamodbav:"created_at" json:"created_at"`
	UpdatedAt   int64  `dynamodbav:"updated_at" json:"updated_at"`
	CompletedAt *int64 `dynamodbav:"completed_at,omitempty" json:"completed_at,omitempty"`
	TTL         int64  `dynamodbav:"ttl" json:"ttl"` // Unix timestamp for auto-deletion (7 days)
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
