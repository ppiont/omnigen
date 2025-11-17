package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/auth"
	"github.com/omnigen/backend/pkg/errors"
	"go.uber.org/zap"
)

// GenerateTitleHandler handles title generation requests
type GenerateTitleHandler struct {
	openAIKey  string
	logger     *zap.Logger
	httpClient *http.Client
}

// NewGenerateTitleHandler creates a new title generation handler
func NewGenerateTitleHandler(openAIKey string, logger *zap.Logger) *GenerateTitleHandler {
	return &GenerateTitleHandler{
		openAIKey: openAIKey,
		logger:    logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateTitleRequest represents a title generation request
type GenerateTitleRequest struct {
	Prompt string `json:"prompt" binding:"required,min=10,max=2000"`
}

// GenerateTitleResponse represents a title generation response
type GenerateTitleResponse struct {
	Title string `json:"title"`
}

// OpenAIRequest represents the OpenAI API request
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the OpenAI API response
type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Choice represents a choice in the OpenAI response
type Choice struct {
	Message Message `json:"message"`
}

// GenerateTitle handles POST /api/v1/generate-title
// @Summary Generate a catchy video title from prompt
// @Description Uses OpenAI GPT-4 to generate a short, engaging video title (max 60 characters)
// @Tags generate
// @Accept json
// @Produce json
// @Param request body GenerateTitleRequest true "Title generation parameters"
// @Success 200 {object} GenerateTitleResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse "Unauthorized"
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/generate-title [post]
// @Security BearerAuth
func (h *GenerateTitleHandler) GenerateTitle(c *gin.Context) {
	var req GenerateTitleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{
			Error: errors.ErrInvalidRequest.WithDetails(map[string]interface{}{
				"validation_error": err.Error(),
			}),
		})
		return
	}

	// Get user ID from auth context (for logging)
	userID := auth.MustGetUserID(c)

	h.logger.Info("Generating title with OpenAI",
		zap.String("user_id", userID),
		zap.String("prompt", req.Prompt[:min(100, len(req.Prompt))]),
	)

	// Check if OpenAI key is available
	if h.openAIKey == "" {
		h.logger.Warn("OpenAI key not configured, using mock title")
		// Return a mock title for local testing
		mockTitle := generateMockTitle(req.Prompt)
		c.JSON(http.StatusOK, GenerateTitleResponse{
			Title: mockTitle,
		})
		return
	}

	// Build OpenAI API request
	systemPrompt := `You are a creative video title generator. Generate a catchy, engaging video title based on the user's prompt. 
The title should be:
- Short and punchy (maximum 60 characters)
- Engaging and attention-grabbing
- Capture the essence of the video content
- Use active language and action words when appropriate
- Avoid generic phrases like "Video about" or "A video showing"

Return ONLY the title text, nothing else. No quotes, no explanations, just the title.`

	openAIReq := OpenAIRequest{
		Model: "gpt-4o",
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Generate a catchy video title for this prompt: %s", req.Prompt),
			},
		},
		Temperature: 0.8,
		MaxTokens:   60,
	}

	payload, err := json.Marshal(openAIReq)
	if err != nil {
		h.logger.Error("Failed to marshal OpenAI request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	// Call OpenAI API
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), "POST",
		"https://api.openai.com/v1/chat/completions",
		bytes.NewReader(payload))
	if err != nil {
		h.logger.Error("Failed to create OpenAI request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	httpReq.Header.Set("Authorization", "Bearer "+h.openAIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		h.logger.Error("OpenAI API request failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer.WithDetails(map[string]interface{}{
				"error": "Failed to connect to OpenAI API",
			}),
		})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("Failed to read OpenAI response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	if resp.StatusCode != http.StatusOK {
		h.logger.Error("OpenAI API error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer.WithDetails(map[string]interface{}{
				"error": fmt.Sprintf("OpenAI API error: %d", resp.StatusCode),
			}),
		})
		return
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		h.logger.Error("Failed to parse OpenAI response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	if openAIResp.Error != nil {
		h.logger.Error("OpenAI API returned error", zap.String("error", openAIResp.Error.Message))
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer.WithDetails(map[string]interface{}{
				"error": openAIResp.Error.Message,
			}),
		})
		return
	}

	if len(openAIResp.Choices) == 0 || openAIResp.Choices[0].Message.Content == "" {
		h.logger.Error("Empty response from OpenAI")
		c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
			Error: errors.ErrInternalServer,
		})
		return
	}

	// Extract and clean title
	title := strings.TrimSpace(openAIResp.Choices[0].Message.Content)
	// Remove quotes if present
	title = strings.Trim(title, `"'`)
	// Ensure max 60 characters
	if len(title) > 60 {
		title = title[:57] + "..."
	}

	h.logger.Info("Title generated successfully",
		zap.String("title", title),
	)

	c.JSON(http.StatusOK, GenerateTitleResponse{
		Title: title,
	})
}

// generateMockTitle generates a simple mock title for local testing
func generateMockTitle(prompt string) string {
	// Extract first few words and create a simple title
	words := strings.Fields(prompt)
	if len(words) == 0 {
		return "Untitled Video"
	}

	// Take first 3-5 words and capitalize
	titleWords := words[:min(5, len(words))]
	title := strings.Join(titleWords, " ")

	// Capitalize first letter of each word
	titleParts := strings.Fields(title)
	for i, part := range titleParts {
		if len(part) > 0 {
			titleParts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	title = strings.Join(titleParts, " ")

	// Ensure max 60 characters
	if len(title) > 60 {
		title = title[:57] + "..."
	}

	return title
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
