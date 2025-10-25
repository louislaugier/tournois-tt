package memes

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MemeGenerator handles video meme creation using Docker
type MemeGenerator struct {
	outputDir         string
	backgroundManager *BackgroundManager
	useBackgrounds    bool
}

// NewMemeGenerator creates a new meme generator
func NewMemeGenerator(outputDir string) *MemeGenerator {
	return &MemeGenerator{
		outputDir:         outputDir,
		backgroundManager: NewBackgroundManager("./meme-templates"),
		useBackgrounds:    true, // Enable backgrounds by default
	}
}

// SetBackgroundsEnabled enables or disables automatic background downloads
func (g *MemeGenerator) SetBackgroundsEnabled(enabled bool) {
	g.useBackgrounds = enabled
}

// GenerateSimpleMeme creates a video with text overlay using Docker + FFmpeg
func (g *MemeGenerator) GenerateSimpleMeme(text string, duration int, outputFilename string) (string, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	outputPath := filepath.Join(g.outputDir, outputFilename)
	absOutputPath, _ := filepath.Abs(outputPath)
	absOutputDir, _ := filepath.Abs(g.outputDir)

	// Escape text for FFmpeg
	escapedText := escapeFFmpegText(text)

	// Docker command to run FFmpeg
	// Creates a solid color background with centered text
	cmd := exec.Command("docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/output", absOutputDir),
		"linuxserver/ffmpeg:latest",
		"-f", "lavfi",
		"-i", fmt.Sprintf("color=c=black:s=1080x1920:d=%d", duration),
		"-vf", fmt.Sprintf(`drawtext=text='%s':fontsize=60:fontcolor=white:x=(w-text_w)/2:y=(h-text_h)/2:fontfile=/usr/share/fonts/dejavu/DejaVuSans-Bold.ttf`, escapedText),
		"-pix_fmt", "yuv420p",
		"-c:v", "libx264",
		fmt.Sprintf("/output/%s", outputFilename),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	return absOutputPath, nil
}

// GenerateMemeWithBackground creates a meme by overlaying text on an existing video
func (g *MemeGenerator) GenerateMemeWithBackground(videoPath, text string, outputFilename string) (string, error) {
	// Ensure output directory exists
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	absVideoPath, _ := filepath.Abs(videoPath)
	absOutputDir, _ := filepath.Abs(g.outputDir)
	videoDir := filepath.Dir(absVideoPath)
	videoFilename := filepath.Base(absVideoPath)

	// Wrap text to fit screen (15 chars per line max)
	wrappedText := wrapTextForFFmpeg(text, 15)

	// Escape only what FFmpeg drawtext needs (not shell escaping)
	// Use %{...} expansion which is more reliable
	ffmpegText := strings.ReplaceAll(wrappedText, "\\", "\\\\")
	ffmpegText = strings.ReplaceAll(ffmpegText, ":", "\\:")
	ffmpegText = strings.ReplaceAll(ffmpegText, "\n", "\\n")
	// For quotes in the text itself, escape them for the shell string
	ffmpegText = strings.ReplaceAll(ffmpegText, "'", "'\"'\"'")

	// Docker command to overlay text on existing video
	cmd := exec.Command("docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/input", videoDir),
		"-v", fmt.Sprintf("%s:/output", absOutputDir),
		"linuxserver/ffmpeg:latest",
		"-i", fmt.Sprintf("/input/%s", videoFilename),
		"-vf", fmt.Sprintf("drawtext=text='%s':fontsize=36:fontcolor=white:x=(w-text_w)/2:y=(h-text_h)/2:box=1:boxcolor=black@0.8:boxborderw=15:line_spacing=6:fontfile=/usr/share/fonts/dejavu/DejaVuSans-Bold.ttf", ffmpegText),
		"-c:a", "copy",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-t", "5",
		fmt.Sprintf("/output/%s", outputFilename),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w\nOutput: %s", err, string(output))
	}

	outputPath := filepath.Join(absOutputDir, outputFilename)
	return outputPath, nil
}

// GenerateMemeWithTemplate generates a meme using a template with matching background
func (g *MemeGenerator) GenerateMemeWithTemplate(template MemeTemplate, outputFilename string) (string, error) {
	// If backgrounds disabled, use simple generation
	if !g.useBackgrounds {
		return g.GenerateSimpleMeme(template.Text, 5, outputFilename)
	}

	// Try to get background for style
	backgroundPath, err := g.backgroundManager.GetBackgroundForStyle(template.Style)
	if err != nil {
		// Fallback to simple meme if background fails
		fmt.Printf("⚠️  Warning: Could not get background for style '%s': %v\n", template.Style, err)
		fmt.Println("   Falling back to simple text meme...")
		return g.GenerateSimpleMeme(template.Text, 5, outputFilename)
	}

	fmt.Printf("📹 Using background: %s\n", filepath.Base(backgroundPath))
	return g.GenerateMemeWithBackground(backgroundPath, template.Text, outputFilename)
}

// wrapTextForFFmpeg wraps text to fit within a pixel width, breaking on words
func wrapTextForFFmpeg(text string, maxCharsPerLine int) string {
	lines := strings.Split(text, "\n")
	var wrappedLines []string

	for _, line := range lines {
		if len(line) <= maxCharsPerLine {
			wrappedLines = append(wrappedLines, line)
			continue
		}

		// Word-wrap this line
		words := strings.Fields(line)
		var currentLine string

		for _, word := range words {
			testLine := currentLine
			if testLine != "" {
				testLine += " "
			}
			testLine += word

			if len(testLine) > maxCharsPerLine {
				if currentLine != "" {
					wrappedLines = append(wrappedLines, currentLine)
					currentLine = word
				} else {
					// Single word too long, just add it
					wrappedLines = append(wrappedLines, word)
					currentLine = ""
				}
			} else {
				currentLine = testLine
			}
		}

		if currentLine != "" {
			wrappedLines = append(wrappedLines, currentLine)
		}
	}

	return strings.Join(wrappedLines, "\n")
}

// escapeFFmpegText escapes special characters for FFmpeg drawtext filter
func escapeFFmpegText(text string) string {
	// Wrap text to fit screen (15 chars per line max)
	text = wrapTextForFFmpeg(text, 15)

	// Escape for FFmpeg drawtext
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, ":", "\\:")
	text = strings.ReplaceAll(text, "\n", "\\n")
	// Escape single quotes for shell
	text = strings.ReplaceAll(text, "'", "'\"'\"'")

	return text
}
