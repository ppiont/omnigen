package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/omnigen/backend/internal/auth"
	backenderrors "github.com/omnigen/backend/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGenerateValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &GenerateHandler{
		logger: zap.NewNop(),
	}

	basePayload := func() map[string]interface{} {
		return map[string]interface{}{
			"prompt":       "Pharmaceutical ad prompt for testing",
			"duration":     30,
			"aspect_ratio": "16:9",
			"voice":        "male",
			"side_effects": strings.Repeat("a", 20),
			"start_image":  "https://example.com/product.png",
		}
	}

	tests := []struct {
		name            string
		mutate          func(map[string]interface{})
		expectedMessage string
		expectedField   string
	}{
		{
			name: "missing voice",
			mutate: func(payload map[string]interface{}) {
				payload["voice"] = ""
			},
			expectedMessage: "Please select a narrator voice (male or female)",
			expectedField:   "voice",
		},
		{
			name: "invalid voice",
			mutate: func(payload map[string]interface{}) {
				payload["voice"] = "robotic"
			},
			expectedMessage: "Invalid voice selection. Choose 'male' or 'female'",
			expectedField:   "voice",
		},
		{
			name: "missing side effects",
			mutate: func(payload map[string]interface{}) {
				payload["side_effects"] = ""
			},
			expectedMessage: "Side effects disclosure is required for pharmaceutical ads",
			expectedField:   "side_effects",
		},
		{
			name: "side effects too short",
			mutate: func(payload map[string]interface{}) {
				payload["side_effects"] = "short"
			},
			expectedMessage: "Side effects text must be at least 10 characters",
			expectedField:   "side_effects",
		},
		{
			name: "side effects too long",
			mutate: func(payload map[string]interface{}) {
				payload["side_effects"] = strings.Repeat("b", 501)
			},
			expectedMessage: "Side effects text cannot exceed 500 characters (currently: 501)",
			expectedField:   "side_effects",
		},
		{
			name: "missing product image",
			mutate: func(payload map[string]interface{}) {
				payload["start_image"] = ""
			},
			expectedMessage: "Product image is required for pharmaceutical ads",
			expectedField:   "start_image",
		},
		{
			name: "duration not achievable with Veo clips",
			mutate: func(payload map[string]interface{}) {
				payload["duration"] = 33
			},
			expectedMessage: "Duration must be between 10-60 seconds and achievable with 4, 6, or 8 second clips",
			expectedField:   "duration",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload := basePayload()
			tc.mutate(payload)

			body, err := json.Marshal(payload)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, err := http.NewRequest(http.MethodPost, "/api/v1/generate", bytes.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			c.Request = req
			c.Set(auth.UserIDKey, "user-123")

			handler.Generate(c)

			require.Equal(t, http.StatusBadRequest, w.Code)

			var resp backenderrors.ErrorResponse
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			require.NotNil(t, resp.Error)
			require.Equal(t, tc.expectedMessage, resp.Error.Message)

			if tc.expectedField != "" {
				require.NotNil(t, resp.Error.Details)
				field, ok := resp.Error.Details["field"].(string)
				require.True(t, ok, "expected field detail to be a string")
				require.Equal(t, tc.expectedField, field)
			}
		})
	}
}
