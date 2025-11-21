package handlers

import (
	"fmt"
	"image/color"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func createTestClip(t *testing.T, width, height int, duration float64) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), fmt.Sprintf("clip_%dx%d.mp4", width, height))
	cmd := exec.Command(
		"ffmpeg",
		"-f", "lavfi",
		"-i", fmt.Sprintf("color=c=black:s=%dx%d:d=%.2f", width, height, duration),
		"-pix_fmt", "yuv420p",
		"-y", path,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to create test clip: %v (%s)", err, strings.TrimSpace(string(output)))
	}
	return path
}

func extractFrame(t *testing.T, videoPath string, timeSec float64) string {
	t.Helper()
	dest := filepath.Join(t.TempDir(), "frame.png")
	cmd := exec.Command(
		"ffmpeg",
		"-ss", fmt.Sprintf("%.2f", timeSec),
		"-i", videoPath,
		"-vframes", "1",
		"-f", "image2",
		"-pix_fmt", "rgba",
		"-y", dest,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to extract frame: %v (%s)", err, strings.TrimSpace(string(output)))
	}
	return dest
}

type frameStats struct {
	width        int
	height       int
	brightPixels int
	minX         int
	maxX         int
	minY         int
	maxY         int
	rowPresence  []int
}

func analyzeFrame(t *testing.T, framePath string, threshold uint32) frameStats {
	t.Helper()
	file, err := os.Open(framePath)
	if err != nil {
		t.Fatalf("failed to open frame: %v", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		t.Fatalf("failed to decode frame: %v", err)
	}

	bounds := img.Bounds()
	stats := frameStats{
		width:  bounds.Dx(),
		height: bounds.Dy(),
		minX:   bounds.Max.X,
		minY:   bounds.Max.Y,
		maxX:   bounds.Min.X,
		maxY:   bounds.Min.Y,
	}

	rowSet := map[int]struct{}{}
	columnCounts := make([]int, bounds.Dx())

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			if rgba.A == 0 {
				continue
			}
			brightness := uint32(rgba.R) + uint32(rgba.G) + uint32(rgba.B)
			if brightness <= threshold {
				continue
			}

			stats.brightPixels++
			if x < stats.minX {
				stats.minX = x
			}
			if x > stats.maxX {
				stats.maxX = x
			}
			if y < stats.minY {
				stats.minY = y
			}
			if y > stats.maxY {
				stats.maxY = y
			}
			rowSet[y] = struct{}{}
			columnCounts[x]++
		}
	}

	if stats.brightPixels == 0 {
		stats.minX, stats.maxX, stats.minY, stats.maxY = -1, -1, -1, -1
		return stats
	}

	const columnThreshold = 5
	newMinX := stats.width
	newMaxX := -1
	for x := 0; x < stats.width; x++ {
		if columnCounts[x] > columnThreshold {
			if x < newMinX {
				newMinX = x
			}
			if x > newMaxX {
				newMaxX = x
			}
		}
	}
	if newMinX < stats.width && newMaxX >= 0 {
		stats.minX = newMinX
		stats.maxX = newMaxX
	}

	for y := range rowSet {
		stats.rowPresence = append(stats.rowPresence, y)
	}
	sort.Ints(stats.rowPresence)

	return stats
}

func (s frameStats) lineCount() int {
	if s.brightPixels == 0 || len(s.rowPresence) == 0 {
		return 0
	}
	const rowGap = 3
	count := 0
	prev := -100
	for _, y := range s.rowPresence {
		if y-prev > rowGap {
			count++
		}
		prev = y
	}
	return count
}

func (s frameStats) centerX() float64 {
	if s.brightPixels == 0 {
		return math.NaN()
	}
	return float64(s.minX+s.maxX) / 2
}

func (s frameStats) widthFraction() float64 {
	if s.brightPixels == 0 {
		return 0
	}
	return float64(s.maxX-s.minX+1) / float64(s.width)
}

func (s frameStats) heightFraction() float64 {
	if s.brightPixels == 0 {
		return 0
	}
	return float64(s.maxY-s.minY+1) / float64(s.height)
}

func (s frameStats) bottomStartFraction() float64 {
	if s.brightPixels == 0 {
		return 0
	}
	return float64(s.minY) / float64(s.height)
}

func ensureFfmpegAvailable(t *testing.T) {
	t.Helper()
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		t.Skip("ffmpeg not available in PATH")
	}
}

func TestComposeVideo_TextOverlayScenarios(t *testing.T) {
	ensureFfmpegAvailable(t)

	logger := zap.NewNop()

	shortText := "May cause drowsiness. Consult your doctor."
	mediumText := "May cause drowsiness, dry mouth, dizziness, or nausea. Do not drive or operate machinery. Avoid alcohol. Consult your doctor if symptoms persist. Not recommended for children under 12."
	longText := "May cause drowsiness, dry mouth, dizziness, nausea, headache, or fatigue. Do not drive or operate heavy machinery. Avoid alcohol consumption. Consult your doctor if symptoms persist or worsen. Not recommended for pregnant women, nursing mothers, or children under 12. May interact with other medications including blood thinners, antidepressants, or sedatives. Stop use and seek medical attention if severe allergic reactions occur including rash, itching, swelling, or difficulty breathing."

	aspectRatios := []struct {
		name   string
		width  int
		height int
	}{
		{"16x9", 1920, 1080},
		{"9x16", 1080, 1920},
		{"1x1", 1080, 1080},
	}

	textScenarios := []struct {
		name           string
		text           string
		expectedLines  int
		maxWidthFactor float64
	}{
		{"short", shortText, 1, 0.65},
		{"medium", mediumText, 2, 0.82},
		{"long", longText, 4, 0.82},
	}

	for _, ratio := range aspectRatios {
		clipPath := createTestClip(t, ratio.width, ratio.height, 5.0)

		for _, scenario := range textScenarios {
			name := fmt.Sprintf("%s_%s", ratio.name, scenario.name)
			t.Run(name, func(t *testing.T) {
				config, err := buildDrawtextConfig(logger, scenario.text, 4.0, 5.0, ratio.width, ratio.height)
				if err != nil {
					t.Fatalf("failed to build drawtext config: %v", err)
				}
				if config == nil {
					t.Fatalf("expected drawtext config for scenario %s", scenario.name)
				}

				finalPath := filepath.Join(t.TempDir(), "final.mp4")
				cmd := exec.Command(
					"ffmpeg",
					"-i", clipPath,
					"-vf", config.Filter,
					"-c:v", "libx264",
					"-preset", "medium",
					"-crf", "21",
					"-y", finalPath,
				)
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("ffmpeg overlay failed: %v (%s)", err, strings.TrimSpace(string(output)))
				}

				preFrame := extractFrame(t, finalPath, 3.0)
				preStats := analyzeFrame(t, preFrame, 80)
				if preStats.brightPixels != 0 {
					t.Fatalf("expected no overlay before side effects start, found %d bright pixels", preStats.brightPixels)
				}

				postFrame := extractFrame(t, finalPath, 4.5)
				postStats := analyzeFrame(t, postFrame, 120)
				if postStats.brightPixels == 0 {
					t.Fatalf("expected overlay in frame, found none")
				}

				bottomStart := postStats.bottomStartFraction()
				if bottomStart < 0.6 {
					t.Fatalf("expected overlay to start near bottom, got %.2f", bottomStart)
				}

				centerX := postStats.centerX()
				centerOffset := math.Abs(centerX - float64(postStats.width)/2.0)
				if centerOffset > float64(postStats.width)*0.12 {
					t.Fatalf("expected text to be horizontally centered, offset %.2f > tolerance", centerOffset)
				}

				maxWidthRatio := scenario.maxWidthFactor
				if ratio.height >= ratio.width {
					switch scenario.name {
					case "short":
						maxWidthRatio = 0.8
					default:
						maxWidthRatio = 0.85
					}
				}
				if config.EstimatedWidth > float64(ratio.width)*maxWidthRatio {
					t.Fatalf("estimated text width exceeds limit: width=%.2fpx limit=%.2fpx (maxChars=%d baseFont=%.2f height=%d)", config.EstimatedWidth, float64(ratio.width)*maxWidthRatio, config.MaxChars, config.BaseFontSize, ratio.height)
				}

				renderedLines := 1
				if strings.TrimSpace(config.RenderedText) != "" {
					renderedLines = strings.Count(config.RenderedText, "\n") + 1
				}

				minLines, maxLines := expectedLineRange(ratio.width, ratio.height, scenario.name)
				if renderedLines < minLines || renderedLines > maxLines {
					t.Fatalf("expected %d-%d lines for %s text, got %d", minLines, maxLines, scenario.name, renderedLines)
				}
			})
		}
	}
}

func expectedLineRange(width, height int, scenario string) (int, int) {
	isVertical := height >= width
	switch scenario {
	case "short":
		if isVertical {
			return 1, 2
		}
		return 1, 1
	case "medium":
		if isVertical {
			return 3, 10
		}
		return 2, 3
	case "long":
		if isVertical {
			return 6, 20
		}
		return 4, 8
	default:
		return 1, 8
	}
}

func TestEscapeFfmpegTextEscapesSpecialCharacters(t *testing.T) {
	input := `It's 100% effective: "Don't miss your dose."`
	want := `It\'s 100\% effective\: \"Don\'t miss your dose.\"`
	if got := escapeFfmpegText(input); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
