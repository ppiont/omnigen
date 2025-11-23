package domain

// BrandGuidelines represents a comprehensive brand guideline document
type BrandGuidelines struct {
	GuidelineID    string    `json:"guideline_id" dynamodbav:"guideline_id"`       // Primary key
	UserID         string    `json:"user_id" dynamodbav:"user_id"`                 // User who uploaded this
	Name           string    `json:"name" dynamodbav:"name"`                       // User-friendly name for the guidelines
	Description    string    `json:"description,omitempty" dynamodbav:"description,omitempty"` // Optional description

	// Visual Brand Elements
	Colors         BrandColors    `json:"colors" dynamodbav:"colors"`               // Primary and secondary colors
	Typography     BrandTypography `json:"typography" dynamodbav:"typography"`      // Font preferences
	LogoAssets     []LogoAsset    `json:"logo_assets" dynamodbav:"logo_assets"`     // Logo variations

	// Tone and Messaging
	BrandVoice     BrandVoice     `json:"brand_voice" dynamodbav:"brand_voice"`     // Tone, personality, messaging

	// Media Guidelines
	ImageStyle     ImageStyle     `json:"image_style" dynamodbav:"image_style"`     // Photography and visual style
	VideoStyle     VideoStyle     `json:"video_style" dynamodbav:"video_style"`     // Video-specific guidelines

	// Document References
	SourceDocument *DocumentRef   `json:"source_document,omitempty" dynamodbav:"source_document,omitempty"` // Original PDF/doc uploaded

	// Metadata
	IsActive       bool          `json:"is_active" dynamodbav:"is_active"`         // Whether this guideline is currently active
	CreatedAt      int64         `json:"created_at" dynamodbav:"created_at"`       // Unix timestamp
	UpdatedAt      int64         `json:"updated_at" dynamodbav:"updated_at"`       // Unix timestamp
}

// BrandColors defines the brand's color palette
type BrandColors struct {
	Primary     []ColorSpec `json:"primary" dynamodbav:"primary"`         // Main brand colors
	Secondary   []ColorSpec `json:"secondary" dynamodbav:"secondary"`     // Supporting colors
	Accent      []ColorSpec `json:"accent" dynamodbav:"accent"`           // Accent/highlight colors
	Neutral     []ColorSpec `json:"neutral" dynamodbav:"neutral"`         // Grays, blacks, whites
	Prohibited  []ColorSpec `json:"prohibited" dynamodbav:"prohibited"`   // Colors to avoid
}

// ColorSpec represents a single color with multiple format representations
type ColorSpec struct {
	Name        string `json:"name" dynamodbav:"name"`               // Color name (e.g., "Primary Blue")
	Hex         string `json:"hex" dynamodbav:"hex"`                 // #FF0000
	RGB         string `json:"rgb,omitempty" dynamodbav:"rgb,omitempty"`     // rgb(255, 0, 0)
	CMYK        string `json:"cmyk,omitempty" dynamodbav:"cmyk,omitempty"`   // For print
	Pantone     string `json:"pantone,omitempty" dynamodbav:"pantone,omitempty"` // Pantone code
	Description string `json:"description,omitempty" dynamodbav:"description,omitempty"` // Usage context
}

// BrandTypography defines font and text styling preferences
type BrandTypography struct {
	Primary     FontSpec   `json:"primary" dynamodbav:"primary"`         // Main brand font
	Secondary   FontSpec   `json:"secondary" dynamodbav:"secondary"`     // Supporting font
	Fallbacks   []string   `json:"fallbacks" dynamodbav:"fallbacks"`     // Web-safe fallback fonts
	Restrictions []string  `json:"restrictions" dynamodbav:"restrictions"` // Fonts to avoid
}

// FontSpec represents a font specification
type FontSpec struct {
	Family       string   `json:"family" dynamodbav:"family"`             // Font family name
	Weights      []string `json:"weights" dynamodbav:"weights"`           // Available weights (light, regular, bold)
	Styles       []string `json:"styles" dynamodbav:"styles"`             // Available styles (italic, normal)
	UsageContext string   `json:"usage_context" dynamodbav:"usage_context"` // When to use this font
}

// LogoAsset represents a logo file reference
type LogoAsset struct {
	AssetID     string `json:"asset_id" dynamodbav:"asset_id"`           // Unique identifier
	Name        string `json:"name" dynamodbav:"name"`                   // Logo variant name
	S3URL       string `json:"s3_url" dynamodbav:"s3_url"`               // S3 storage location
	ContentType string `json:"content_type" dynamodbav:"content_type"`   // MIME type
	FileSize    int64  `json:"file_size" dynamodbav:"file_size"`         // File size in bytes
	Variant     string `json:"variant" dynamodbav:"variant"`             // horizontal, vertical, icon, etc.
	ColorType   string `json:"color_type" dynamodbav:"color_type"`       // full-color, monochrome, white, etc.
	Usage       string `json:"usage" dynamodbav:"usage"`                 // Usage guidelines
}

// BrandVoice defines the brand's personality and messaging approach
type BrandVoice struct {
	Personality    []string `json:"personality" dynamodbav:"personality"`       // Brand personality traits
	ToneKeywords   []string `json:"tone_keywords" dynamodbav:"tone_keywords"`   // Descriptive tone words
	MessagingStyle string   `json:"messaging_style" dynamodbav:"messaging_style"` // formal, casual, playful, etc.
	DoList         []string `json:"do_list" dynamodbav:"do_list"`               // Things to do in messaging
	DontList       []string `json:"dont_list" dynamodbav:"dont_list"`           // Things to avoid
	SampleMessages []string `json:"sample_messages" dynamodbav:"sample_messages"` // Example brand messages
}

// ImageStyle defines photography and visual content guidelines
type ImageStyle struct {
	Style          string   `json:"style" dynamodbav:"style"`                   // Photography style
	Subjects       []string `json:"subjects" dynamodbav:"subjects"`             // Preferred subjects/themes
	Composition    []string `json:"composition" dynamodbav:"composition"`       // Composition preferences
	ColorTreatment string   `json:"color_treatment" dynamodbav:"color_treatment"` // Color processing style
	Mood           string   `json:"mood" dynamodbav:"mood"`                     // Overall mood/feeling
	Restrictions   []string `json:"restrictions" dynamodbav:"restrictions"`     // What to avoid in images
}

// VideoStyle defines video content guidelines
type VideoStyle struct {
	PacingStyle    string   `json:"pacing_style" dynamodbav:"pacing_style"`       // fast, medium, slow
	TransitionType []string `json:"transition_type" dynamodbav:"transition_type"` // Preferred transitions
	CameraMovement []string `json:"camera_movement" dynamodbav:"camera_movement"` // Preferred camera movements
	ColorGrading   string   `json:"color_grading" dynamodbav:"color_grading"`     // Color grading style
	TextOverlays   TextOverlayStyle `json:"text_overlays" dynamodbav:"text_overlays"` // Text overlay guidelines
}

// TextOverlayStyle defines how text should appear in videos
type TextOverlayStyle struct {
	FontFamily     string `json:"font_family" dynamodbav:"font_family"`         // Preferred font for overlays
	FontSize       string `json:"font_size" dynamodbav:"font_size"`             // Size guidelines (large, medium, small)
	Color          string `json:"color" dynamodbav:"color"`                     // Text color (often from brand colors)
	Position       string `json:"position" dynamodbav:"position"`               // Positioning preference
	Animation      string `json:"animation" dynamodbav:"animation"`             // Animation style for text
	Background     string `json:"background" dynamodbav:"background"`           // Background treatment for text
}

// DocumentRef represents a reference to an uploaded brand document
type DocumentRef struct {
	S3URL       string `json:"s3_url" dynamodbav:"s3_url"`               // S3 storage location
	Filename    string `json:"filename" dynamodbav:"filename"`           // Original filename
	ContentType string `json:"content_type" dynamodbav:"content_type"`   // MIME type
	FileSize    int64  `json:"file_size" dynamodbav:"file_size"`         // File size in bytes
	UploadedAt  int64  `json:"uploaded_at" dynamodbav:"uploaded_at"`     // Unix timestamp
}

// BrandGuidelineRequest represents the request to create/update brand guidelines
type BrandGuidelineRequest struct {
	Name           string                 `json:"name" binding:"required"`
	Description    string                 `json:"description,omitempty"`
	Colors         *BrandColors           `json:"colors,omitempty"`
	Typography     *BrandTypography       `json:"typography,omitempty"`
	LogoAssets     []LogoAsset            `json:"logo_assets,omitempty"`
	BrandVoice     *BrandVoice            `json:"brand_voice,omitempty"`
	ImageStyle     *ImageStyle            `json:"image_style,omitempty"`
	VideoStyle     *VideoStyle            `json:"video_style,omitempty"`
	SourceDocument *DocumentRef           `json:"source_document,omitempty"`

	// For document parsing - client can upload a document and we'll extract guidelines
	ParseFromDocument bool `json:"parse_from_document,omitempty"`
}

// BrandGuidelineSummary provides a condensed version for API responses
type BrandGuidelineSummary struct {
	GuidelineID    string    `json:"guideline_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	IsActive       bool      `json:"is_active"`
	HasColors      bool      `json:"has_colors"`
	HasLogos       bool      `json:"has_logos"`
	HasVoice       bool      `json:"has_voice"`
	HasImageStyle  bool      `json:"has_image_style"`
	HasVideoStyle  bool      `json:"has_video_style"`
	CreatedAt      int64     `json:"created_at"`
	UpdatedAt      int64     `json:"updated_at"`
}

// ToSummary converts a full BrandGuidelines to a summary
func (bg *BrandGuidelines) ToSummary() BrandGuidelineSummary {
	return BrandGuidelineSummary{
		GuidelineID:    bg.GuidelineID,
		Name:           bg.Name,
		Description:    bg.Description,
		IsActive:       bg.IsActive,
		HasColors:      len(bg.Colors.Primary) > 0 || len(bg.Colors.Secondary) > 0,
		HasLogos:       len(bg.LogoAssets) > 0,
		HasVoice:       len(bg.BrandVoice.Personality) > 0,
		HasImageStyle:  bg.ImageStyle.Style != "",
		HasVideoStyle:  bg.VideoStyle.PacingStyle != "",
		CreatedAt:      bg.CreatedAt,
		UpdatedAt:      bg.UpdatedAt,
	}
}

// ForVideoGeneration creates a simplified string representation suitable for AI video generation prompts
func (bg *BrandGuidelines) ForVideoGeneration() string {
	if bg == nil {
		return ""
	}

	var parts []string

	// Add color information
	if len(bg.Colors.Primary) > 0 {
		colorNames := make([]string, len(bg.Colors.Primary))
		for i, color := range bg.Colors.Primary {
			colorNames[i] = color.Name + " (" + color.Hex + ")"
		}
		parts = append(parts, "Primary colors: "+joinStrings(colorNames, ", "))
	}

	// Add brand voice
	if len(bg.BrandVoice.Personality) > 0 {
		parts = append(parts, "Brand personality: "+joinStrings(bg.BrandVoice.Personality, ", "))
	}
	if bg.BrandVoice.MessagingStyle != "" {
		parts = append(parts, "Messaging style: "+bg.BrandVoice.MessagingStyle)
	}

	// Add visual style
	if bg.ImageStyle.Style != "" {
		parts = append(parts, "Visual style: "+bg.ImageStyle.Style)
	}
	if bg.ImageStyle.Mood != "" {
		parts = append(parts, "Mood: "+bg.ImageStyle.Mood)
	}

	// Add video-specific guidelines
	if bg.VideoStyle.PacingStyle != "" {
		parts = append(parts, "Video pacing: "+bg.VideoStyle.PacingStyle)
	}
	if bg.VideoStyle.ColorGrading != "" {
		parts = append(parts, "Color grading: "+bg.VideoStyle.ColorGrading)
	}

	if len(parts) == 0 {
		return ""
	}

	return "Brand guidelines: " + joinStrings(parts, "; ")
}

// Helper function to join strings (Go doesn't have this in stdlib for []string)
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}