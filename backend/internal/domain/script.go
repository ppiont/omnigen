package domain

// Script represents a complete ad creative script with production-ready scene specifications
type Script struct {
	ScriptID         string    `json:"script_id" dynamodbav:"script_id"`
	UserID           string    `json:"user_id" dynamodbav:"user_id"`
	Title            string    `json:"title" dynamodbav:"title"`
	TotalDuration    int       `json:"total_duration" dynamodbav:"total_duration"` // seconds
	Scenes           []Scene   `json:"scenes" dynamodbav:"scenes"`
	AudioSpec        AudioSpec `json:"audio_spec" dynamodbav:"audio_spec"`
	Metadata         Metadata  `json:"metadata" dynamodbav:"metadata"`
	StyleDescription string    `json:"style_description,omitempty" dynamodbav:"style_description,omitempty"` // Extracted from style reference image
	CreatedAt        int64     `json:"created_at" dynamodbav:"created_at"`                                   // Unix timestamp
	UpdatedAt        int64     `json:"updated_at" dynamodbav:"updated_at"`                                   // Unix timestamp
	Status           string    `json:"status" dynamodbav:"status"`                                           // "draft", "approved", "generating", "completed"
	ExpiresAt        int64     `json:"expires_at,omitempty" dynamodbav:"expires_at,omitempty"`               // TTL timestamp
}

// Scene represents a single shot/scene in the advertisement with cinematography details
type Scene struct {
	SceneNumber int     `json:"scene_number"`
	StartTime   float64 `json:"start_time"` // seconds from start
	Duration    float64 `json:"duration"`   // scene length in seconds

	// Screenplay elements
	Location string `json:"location"` // e.g., "INT. COFFEE SHOP - DAY", "EXT. BEACH - GOLDEN HOUR"
	Action   string `json:"action"`   // What happens in this scene (2-3 sentences)

	// Visual Direction (Cinematography)
	ShotType    ShotType    `json:"shot_type"`    // Wide, medium, close-up, etc.
	CameraAngle CameraAngle `json:"camera_angle"` // Eye-level, high angle, low angle, etc.
	CameraMove  CameraMove  `json:"camera_move"`  // Static, dolly, pan, tilt, etc.

	// Lighting & Mood
	Lighting   Lighting   `json:"lighting"`    // Natural, studio, dramatic, etc.
	ColorGrade ColorGrade `json:"color_grade"` // Warm, cool, desaturated, etc.
	Mood       Mood       `json:"mood"`        // Energetic, calm, dramatic, etc.

	// Style & Aesthetic
	VisualStyle VisualStyle `json:"visual_style"` // Cinematic, documentary, minimalist, etc.

	// Transitions
	TransitionIn  Transition `json:"transition_in"`  // How to enter this scene
	TransitionOut Transition `json:"transition_out"` // How to exit this scene

	// AI Generation
	GenerationPrompt string `json:"generation_prompt"`         // Optimized prompt for Veo 3.1
	StartImageURL    string `json:"start_image_url,omitempty"` // For visual continuity between scenes
}

// AudioSpec defines the audio requirements for the advertisement
type AudioSpec struct {
	EnableAudio   bool   `json:"enable_audio"`
	MusicMood     string `json:"music_mood"`               // "upbeat", "calm", "dramatic", etc.
	MusicStyle    string `json:"music_style"`              // "electronic", "acoustic", "orchestral", etc.
	VoiceoverText string `json:"voiceover_text,omitempty"` // Optional voiceover script (legacy)

	// Pharmaceutical ad narrator metadata
	NarratorScript       string  `json:"narrator_script,omitempty"`         // Full narrator script including side effects
	SideEffectsText      string  `json:"side_effects_text,omitempty"`       // Exact text for on-screen disclosure
	SideEffectsStartTime float64 `json:"side_effects_start_time,omitempty"` // Timestamp (seconds) when side effects begin

	SyncPoints []SyncPoint `json:"sync_points"` // Audio-visual synchronization markers
}

// SyncPoint marks specific audio-visual synchronization moments
type SyncPoint struct {
	Timestamp   float64 `json:"timestamp"` // Time in seconds
	Type        string  `json:"type"`      // "beat", "voiceover", "sfx", "transition"
	SceneNumber int     `json:"scene_number"`
	Description string  `json:"description"` // e.g., "Product reveal on beat drop"
}

// Metadata contains creative direction and brand information
type Metadata struct {
	ProductName    string   `json:"product_name"`
	BrandGuideline string   `json:"brand_guideline,omitempty"` // Brand colors, tone, values
	TargetAudience string   `json:"target_audience"`
	CallToAction   string   `json:"call_to_action"` // Final message/CTA
	Keywords       []string `json:"keywords"`       // Key themes/concepts
}

// --- CINEMATOGRAPHY ENUMS (Industry-Standard Terminology) ---

// ShotType defines the framing of the shot
type ShotType string

const (
	ShotExtremeWide  ShotType = "extreme_wide_shot"  // Establishing shot, shows environment
	ShotWide         ShotType = "wide_shot"          // Shows full subject and surroundings
	ShotFull         ShotType = "full_shot"          // Subject fills frame head to toe
	ShotCowboy       ShotType = "cowboy_shot"        // Mid-thighs up (American shot)
	ShotMedium       ShotType = "medium_shot"        // Waist up
	ShotMediumClose  ShotType = "medium_close_up"    // Chest up
	ShotCloseUp      ShotType = "close_up"           // Head and shoulders
	ShotExtremeClose ShotType = "extreme_close_up"   // Eyes, hands, product details
	ShotOverShoulder ShotType = "over_shoulder_shot" // OTS for dialogue/interaction
	ShotTwoShot      ShotType = "two_shot"           // Two subjects in frame
	ShotInsert       ShotType = "insert_shot"        // Detail shot of object/action
)

// CameraAngle defines the vertical position of the camera
type CameraAngle string

const (
	AngleEyeLevel CameraAngle = "eye_level"      // Standard, neutral perspective
	AngleHigh     CameraAngle = "high_angle"     // Camera above subject (diminutive)
	AngleLow      CameraAngle = "low_angle"      // Camera below subject (powerful)
	AngleDutch    CameraAngle = "dutch_angle"    // Tilted frame (tension, unease)
	AngleBirdsEye CameraAngle = "birds_eye"      // Directly overhead (map view)
	AngleWorms    CameraAngle = "worms_eye"      // Directly below looking up
	AngleShoulder CameraAngle = "shoulder_level" // Over-the-shoulder height
)

// CameraMove defines camera movement during the shot
type CameraMove string

const (
	MoveStatic     CameraMove = "static"       // Locked-off camera, no movement
	MovePanLeft    CameraMove = "pan_left"     // Horizontal rotation left
	MovePanRight   CameraMove = "pan_right"    // Horizontal rotation right
	MoveTiltUp     CameraMove = "tilt_up"      // Vertical rotation upward
	MoveTiltDown   CameraMove = "tilt_down"    // Vertical rotation downward
	MoveDollyIn    CameraMove = "dolly_in"     // Push in toward subject
	MoveDollyOut   CameraMove = "dolly_out"    // Pull back from subject
	MoveDollyLeft  CameraMove = "dolly_left"   // Lateral movement left (truck left)
	MoveDollyRight CameraMove = "dolly_right"  // Lateral movement right (truck right)
	MoveZoomIn     CameraMove = "zoom_in"      // Lens zoom in (not dolly)
	MoveZoomOut    CameraMove = "zoom_out"     // Lens zoom out
	MoveHandheld   CameraMove = "handheld"     // Organic, shaky movement
	MoveSteadycam  CameraMove = "steadycam"    // Smooth handheld
	MoveArc        CameraMove = "arc"          // Circular movement around subject
	MoveTracking   CameraMove = "tracking"     // Follow subject movement
	MoveCrane      CameraMove = "crane_up"     // Vertical crane movement up
	MoveCraneDown  CameraMove = "crane_down"   // Vertical crane movement down
	MoveDrone      CameraMove = "drone_aerial" // Aerial drone movement
)

// Lighting defines the lighting setup and quality
type Lighting string

const (
	LightNatural    Lighting = "natural_light"      // Daylight, window light
	LightGoldenHour Lighting = "golden_hour"        // Sunrise/sunset warm light
	LightBlueHour   Lighting = "blue_hour"          // Twilight, cool ambient
	LightStudio     Lighting = "studio_lighting"    // Controlled 3-point lighting
	LightDramatic   Lighting = "dramatic_lighting"  // High contrast, shadows
	LightSoft       Lighting = "soft_lighting"      // Diffused, even illumination
	LightHardLight  Lighting = "hard_light"         // Direct, sharp shadows
	LightBacklit    Lighting = "backlit"            // Light from behind subject
	LightRimLight   Lighting = "rim_lighting"       // Edge lighting
	LightLowKey     Lighting = "low_key"            // Dark, moody (film noir)
	LightHighKey    Lighting = "high_key"           // Bright, minimal shadows
	LightNeon       Lighting = "neon_lighting"      // Artificial neon/LED colors
	LightPractical  Lighting = "practical_lighting" // In-scene light sources
	LightSilhouette Lighting = "silhouette"         // Subject as dark shape
)

// ColorGrade defines the color treatment/grading style
type ColorGrade string

const (
	GradeNatural     ColorGrade = "natural"       // Realistic, true-to-life colors
	GradeWarm        ColorGrade = "warm_tones"    // Orange/yellow push (cozy)
	GradeCool        ColorGrade = "cool_tones"    // Blue/teal push (modern)
	GradeTealOrange  ColorGrade = "teal_orange"   // Hollywood blockbuster look
	GradeDesaturated ColorGrade = "desaturated"   // Muted, washed-out colors
	GradeVibrant     ColorGrade = "vibrant"       // Saturated, punchy colors
	GradeMonochrome  ColorGrade = "monochrome"    // Black and white
	GradeSepia       ColorGrade = "sepia"         // Vintage brown tone
	GradeBleach      ColorGrade = "bleach_bypass" // High contrast, low saturation
	GradeCinematic   ColorGrade = "cinematic"     // Film-like contrast curves
	GradePastel      ColorGrade = "pastel"        // Soft, muted pastels
	GradeNoir        ColorGrade = "noir"          // High contrast B&W
	GradeRetro       ColorGrade = "retro_film"    // Vintage film emulation
)

// Mood defines the emotional tone of the scene
type Mood string

const (
	MoodEnergetic     Mood = "energetic"     // High energy, exciting
	MoodCalm          Mood = "calm"          // Peaceful, serene
	MoodDramatic      Mood = "dramatic"      // Intense, emotional
	MoodInspiring     Mood = "inspiring"     // Uplifting, motivational
	MoodMysterious    Mood = "mysterious"    // Enigmatic, intriguing
	MoodPlayful       Mood = "playful"       // Fun, lighthearted
	MoodSophisticated Mood = "sophisticated" // Elegant, refined
	MoodNostalgic     Mood = "nostalgic"     // Sentimental, reminiscent
	MoodUrgent        Mood = "urgent"        // Fast-paced, pressing
	MoodLuxurious     Mood = "luxurious"     // Premium, high-end
	MoodIntimate      Mood = "intimate"      // Close, personal
	MoodEpic          Mood = "epic"          // Grand, awe-inspiring
)

// VisualStyle defines the overall aesthetic approach
type VisualStyle string

const (
	StyleCinematic   VisualStyle = "cinematic"       // Film-like, narrative driven
	StyleDocumentary VisualStyle = "documentary"     // Realistic, authentic
	StyleMinimalist  VisualStyle = "minimalist"      // Clean, simple, uncluttered
	StyleMaximalist  VisualStyle = "maximalist"      // Rich, detailed, layered
	StyleCommercial  VisualStyle = "commercial"      // Polished, advertising aesthetic
	StyleEditorial   VisualStyle = "editorial"       // Fashion magazine style
	StyleLifestyle   VisualStyle = "lifestyle"       // Everyday, relatable
	StyleProduct     VisualStyle = "product_focused" // Product hero shots
	StyleAbstract    VisualStyle = "abstract"        // Artistic, conceptual
	StyleVintage     VisualStyle = "vintage"         // Retro, period-specific
	StyleFuturistic  VisualStyle = "futuristic"      // Modern, tech-forward
	StyleGritty      VisualStyle = "gritty"          // Raw, textured
	StyleDreamy      VisualStyle = "dreamy"          // Soft focus, ethereal
)

// Transition defines how to move between scenes
type Transition string

const (
	TransitionCut       Transition = "cut"             // Instant change (standard)
	TransitionFade      Transition = "fade"            // Fade to black
	TransitionCrossFade Transition = "cross_fade"      // Dissolve between scenes
	TransitionWipeLeft  Transition = "wipe_left"       // Wipe transition left
	TransitionWipeRight Transition = "wipe_right"      // Wipe transition right
	TransitionIrisIn    Transition = "iris_in"         // Circular reveal
	TransitionIrisOut   Transition = "iris_out"        // Circular close
	TransitionMatchCut  Transition = "match_cut"       // Visual/motion match
	TransitionJumpCut   Transition = "jump_cut"        // Jarring time skip
	TransitionSmashCut  Transition = "smash_cut"       // Abrupt contrast cut
	TransitionWhip      Transition = "whip_pan"        // Fast camera whip
	TransitionZoom      Transition = "zoom_transition" // Zoom blur
	TransitionNone      Transition = "none"            // First/last scene
)
