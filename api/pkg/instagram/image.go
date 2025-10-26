package instagram

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tournois-tt/api/pkg/utils"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// Modern color palette for Instagram
var (
	// Gradient colors
	ColorGradientStart = color.RGBA{31, 186, 214, 255} // #1FBAD6 - Turquoise
	ColorGradientEnd   = color.RGBA{155, 89, 182, 255} // #9B59B6 - Purple

	// Clean backgrounds
	ColorWhite     = color.RGBA{255, 255, 255, 255}
	ColorLightGray = color.RGBA{248, 249, 250, 255}
	ColorDarkGray  = color.RGBA{73, 80, 87, 255}
	ColorBlack     = color.RGBA{33, 37, 41, 255}

	// Accent colors
	ColorOrange  = color.RGBA{255, 127, 80, 255} // Coral orange for emphasis
	ColorSuccess = color.RGBA{40, 167, 69, 255}  // Green for positive info
)

const (
	// Instagram optimal size (square)
	ImageWidth  = 1080
	ImageHeight = 1080
)

var (
	regularFont *opentype.Font
	boldFont    *opentype.Font
)

func init() {
	var err error
	regularFont, err = opentype.Parse(goregular.TTF)
	if err != nil {
		panic(err)
	}
	boldFont, err = opentype.Parse(gobold.TTF)
	if err != nil {
		panic(err)
	}
}

// GenerateTournamentImage creates a modern, minimalistic tournament image for Instagram
func GenerateTournamentImage(tournamentData TournamentImage) (string, error) {
	// Create canvas
	img := image.NewRGBA(image.Rect(0, 0, ImageWidth, ImageHeight))

	// Fill with clean white background
	draw.Draw(img, img.Bounds(), &image.Uniform{ColorWhite}, image.Point{}, draw.Src)

	// Add subtle gradient accent bar at top
	gradientHeight := 12
	drawGradientBar(img, 0, gradientHeight)

	y := 65

	// Header section - centered
	y = drawCenteredText(img, "Nouvelle homologation", y, ColorBlack, 46, boldFont)
	y += 50

	// Tournament name (biggest text, bold, in quotes, centered)
	tournamentName := wrapText(tournamentData.Name, 32)
	tournamentNameWithQuotes := fmt.Sprintf("\"%s\"", tournamentName)
	y = drawCenteredText(img, tournamentNameWithQuotes, y, ColorBlack, 44, boldFont)

	y += 40

	// Tournament type badge (centered)
	mappedType := utils.MapTournamentType(tournamentData.Type)
	y = drawCenteredBadge(img, mappedType, y, ColorGradientStart)
	y += 40

	// Key info - centered (with tighter spacing to prevent bottom cutoff)
	if tournamentData.Endowment > 0 {
		y = drawCenteredInfoLine(img, "DOTATION TOTALE", fmt.Sprintf("%d €", tournamentData.Endowment/100), y)
		y += 40
	}

	dateStr := formatDates(tournamentData.StartDate, tournamentData.EndDate)
	// Use "DATE" for single day, "DATES" for multiple days
	dateLabel := "DATE"
	if tournamentData.StartDate != tournamentData.EndDate {
		dateLabel = "DATES"
	}
	y = drawCenteredInfoLine(img, dateLabel, dateStr, y)
	y += 40

	clubName := wrapText(tournamentData.Club, 38)
	y = drawCenteredInfoLine(img, "CLUB ORGANISATEUR", clubName, y)
	y += 40

	address := wrapText(tournamentData.Address, 38)
	y = drawCenteredInfoLine(img, "LIEU", address, y)
	y += 40

	// Footer with URL - centered (ensure enough bottom margin)
	footerY := y

	// Subtle separator
	drawSeparator(img, 60, ImageWidth-60, footerY)
	footerY += 20

	// "Règlement" label
	footerY = drawCenteredText(img, "RÈGLEMENT", footerY, ColorDarkGray, 20, boldFont)
	footerY += 7

	// URL in accent color (centered) - smaller font to ensure it fits
	_ = drawCenteredText(img, tournamentData.TournamentURL, footerY, ColorGradientStart, 24, boldFont)

	// Save to instagram-images folder
	imagesDir := "./instagram-images"
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create images directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("tournament_%d_%s.png", tournamentData.TournamentID, timestamp)
	filePath := filepath.Join(imagesDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	return filePath, nil
}

// drawGradientBar draws a horizontal gradient bar
func drawGradientBar(img *image.RGBA, y, height int) {
	for row := y; row < y+height; row++ {
		for col := 0; col < ImageWidth; col++ {
			// Simple linear gradient
			ratio := float64(col) / float64(ImageWidth)
			r := uint8(float64(ColorGradientStart.R)*(1-ratio) + float64(ColorGradientEnd.R)*ratio)
			g := uint8(float64(ColorGradientStart.G)*(1-ratio) + float64(ColorGradientEnd.G)*ratio)
			b := uint8(float64(ColorGradientStart.B)*(1-ratio) + float64(ColorGradientEnd.B)*ratio)
			img.Set(col, row, color.RGBA{r, g, b, 255})
		}
	}
}

// drawTextWithFont draws text with a TrueType font at specified size
func drawTextWithFont(img *image.RGBA, text string, x, y int, col color.RGBA, size float64, ttfFont *opentype.Font) int {
	face, err := opentype.NewFace(ttfFont, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return y
	}
	defer face.Close()

	drawer := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{col},
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}

	lines := strings.Split(text, "\n")
	metrics := face.Metrics()
	lineHeight := metrics.Height.Ceil() + int(size*0.4) // Add more line spacing

	currentY := y
	for i, line := range lines {
		if i > 0 {
			currentY += lineHeight
			drawer.Dot = fixed.Point26_6{X: fixed.I(x), Y: fixed.I(currentY)}
		}
		drawer.DrawString(line)
	}

	// Return Y position after all lines
	return currentY + lineHeight
}

// drawInfoLine draws an info line with emoji and text
func drawInfoLine(img *image.RGBA, emoji, text string, x, y int) int {
	// Draw emoji
	drawTextWithFont(img, emoji, x, y, ColorBlack, 44, regularFont)

	// Draw text next to emoji (handle multiline properly)
	endY := drawTextWithFont(img, text, x+80, y, ColorBlack, 42, regularFont)
	return endY
}

// drawInfoLineWithLabel draws an info line with a text label (instead of emoji)
func drawInfoLineWithLabel(img *image.RGBA, label, text string, x, y int) int {
	// Draw label in smaller, uppercase, gray text
	drawTextWithFont(img, label, x, y, ColorDarkGray, 20, boldFont)

	// Draw main text below the label (more spacing for better readability)
	endY := drawTextWithFont(img, text, x, y+32, ColorBlack, 32, regularFont)
	return endY
}

// drawBadge draws a colored badge with text
func drawBadge(img *image.RGBA, text string, x, y int, bgColor color.RGBA) int {
	// Get text dimensions
	face, _ := opentype.NewFace(boldFont, &opentype.FaceOptions{
		Size:    26,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	defer face.Close()

	// Measure text width more accurately
	drawer := &font.Drawer{Face: face}
	textWidth := drawer.MeasureString(text).Ceil()

	// Draw rounded rectangle background with proper padding
	padding := 18
	badgeWidth := textWidth + (padding * 2)
	badgeHeight := 42

	// Fill badge background
	for row := y - 32; row < y-32+badgeHeight; row++ {
		for col := x; col < x+badgeWidth && col < ImageWidth; col++ {
			img.Set(col, row, bgColor)
		}
	}

	// Draw white text on badge (centered with padding)
	return drawTextWithFont(img, text, x+padding, y, ColorWhite, 26, boldFont)
}

// drawSeparator draws a thin horizontal line
func drawSeparator(img *image.RGBA, x1, x2, y int) {
	for x := x1; x < x2; x++ {
		img.Set(x, y, ColorLightGray)
		img.Set(x, y+1, ColorLightGray)
	}
}

// wrapText wraps text to fit within a certain character width
func wrapText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}

	words := strings.Fields(text)
	var lines []string
	var currentLine string

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) > maxWidth {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				lines = append(lines, word)
				currentLine = ""
			}
		} else {
			currentLine = testLine
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

// drawCenteredText draws centered text on the image
func drawCenteredText(img *image.RGBA, text string, y int, col color.RGBA, size float64, ttfFont *opentype.Font) int {
	face, err := opentype.NewFace(ttfFont, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return y
	}
	defer face.Close()

	drawer := &font.Drawer{Face: face}
	lines := strings.Split(text, "\n")
	metrics := face.Metrics()
	lineHeight := metrics.Height.Ceil() + int(size*0.4)

	currentY := y
	for i, line := range lines {
		if i > 0 {
			currentY += lineHeight
		}

		// Measure line width and center it
		textWidth := drawer.MeasureString(line).Ceil()
		x := (ImageWidth - textWidth) / 2

		// Draw the line
		drawer.Dst = img
		drawer.Src = &image.Uniform{col}
		drawer.Dot = fixed.Point26_6{X: fixed.I(x), Y: fixed.I(currentY)}
		drawer.DrawString(line)
	}

	return currentY + lineHeight
}

// drawCenteredInfoLine draws a centered info line with label and text
func drawCenteredInfoLine(img *image.RGBA, label, text string, y int) int {
	// Draw label centered
	y = drawCenteredText(img, label, y, ColorDarkGray, 20, boldFont)

	// Draw main text centered below the label
	endY := drawCenteredText(img, text, y+7, ColorBlack, 32, regularFont)
	return endY
}

// drawCenteredBadge draws a centered colored badge with text
func drawCenteredBadge(img *image.RGBA, text string, y int, bgColor color.RGBA) int {
	// Get text dimensions
	face, _ := opentype.NewFace(boldFont, &opentype.FaceOptions{
		Size:    26,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	defer face.Close()

	// Measure text width
	drawer := &font.Drawer{Face: face}
	textWidth := drawer.MeasureString(text).Ceil()

	// Calculate centered position
	padding := 18
	badgeWidth := textWidth + (padding * 2)
	badgeHeight := 42
	x := (ImageWidth - badgeWidth) / 2

	// Fill badge background
	for row := y - 32; row < y-32+badgeHeight; row++ {
		for col := x; col < x+badgeWidth && col < ImageWidth; col++ {
			img.Set(col, row, bgColor)
		}
	}

	// Draw white text on badge (centered)
	return drawCenteredText(img, text, y, ColorWhite, 26, boldFont)
}

// formatDates formats start and end dates for display
func formatDates(startDate, endDate string) string {
	if startDate == "" {
		return "Date non disponible"
	}

	// Parse dates (handle multiple formats)
	var start time.Time
	var err error

	// Try RFC3339 first
	start, err = time.Parse(time.RFC3339, startDate)
	if err != nil {
		// Try without timezone (common format from backend)
		start, err = time.Parse("2006-01-02T15:04:05", startDate)
		if err != nil {
			// Try simple date format
			start, err = time.Parse("2006-01-02", startDate)
			if err != nil {
				return startDate // Return as-is if all parsing fails
			}
		}
	}

	// If no end date or same as start date, show single date
	if endDate == "" || endDate == startDate {
		return start.Format("02/01/2006")
	}

	var end time.Time
	// Try same parsing logic for end date
	end, err = time.Parse(time.RFC3339, endDate)
	if err != nil {
		end, err = time.Parse("2006-01-02T15:04:05", endDate)
		if err != nil {
			end, err = time.Parse("2006-01-02", endDate)
			if err != nil {
				return start.Format("02/01/2006") // Just show start date if end parsing fails
			}
		}
	}

	// If same month, show range within month
	if start.Month() == end.Month() && start.Year() == end.Year() {
		return fmt.Sprintf("%d-%d %s %d", start.Day(), end.Day(), monthName(start.Month()), start.Year())
	}

	// Otherwise show full range
	return fmt.Sprintf("%s - %s", start.Format("02/01/2006"), end.Format("02/01/2006"))
}

// monthName returns the French month name
func monthName(m time.Month) string {
	months := []string{
		"janvier", "février", "mars", "avril", "mai", "juin",
		"juillet", "août", "septembre", "octobre", "novembre", "décembre",
	}
	return months[m-1]
}

// CleanupImage removes a temporary tournament image file
func CleanupImage(filepath string) error {
	return os.Remove(filepath)
}
