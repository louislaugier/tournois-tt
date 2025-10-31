package instagram

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
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
	scales := []float64{1.0, 0.95, 0.9, 0.85, 0.8, 0.75}
	maxContentHeight := ImageHeight - 80

	var finalImg *image.RGBA
	finalY := math.MaxInt32

	for idx, scale := range scales {
		img := createBaseImage()
		bottom := renderTournamentImage(img, tournamentData, scale)

		if bottom < finalY {
			finalImg = img
			finalY = bottom
		}

		if bottom <= maxContentHeight {
			finalImg = img
			finalY = bottom
			break
		}

		// Ensure we keep at least the last attempt even if it doesn't fit perfectly
		if idx == len(scales)-1 && finalImg == nil {
			finalImg = img
			finalY = bottom
		}
	}

	if finalImg == nil {
		return "", fmt.Errorf("failed to render tournament image")
	}

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

	if err := png.Encode(file, finalImg); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	_ = finalY // retained for future diagnostics if needed

	return filePath, nil
}

func createBaseImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, ImageWidth, ImageHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{ColorWhite}, image.Point{}, draw.Src)
	drawGradientBar(img, 0, 12)
	return img
}

func renderTournamentImage(img *image.RGBA, tournamentData TournamentImage, scale float64) int {
	y := scaledSpacing(80, scale, 48)

	headerSize := scaledFontSize(46, scale, 28)
	y = drawCenteredText(img, "Nouvelle homologation (FFTT)", y, ColorBlack, headerSize, boldFont)
	y += scaledSpacing(40, scale, 24)

	tournamentName := wrapText(tournamentData.Name, 32, scale)
	tournamentNameWithQuotes := fmt.Sprintf("\"%s\"", tournamentName)
	nameSize := scaledFontSize(44, scale, 26)
	y = drawCenteredText(img, tournamentNameWithQuotes, y, ColorBlack, nameSize, boldFont)
	y += scaledSpacing(32, scale, 22)

	mappedType := utils.MapTournamentType(tournamentData.Type)
	y = drawCenteredBadge(img, mappedType, y, ColorGradientStart, scale)
	y += scaledSpacing(28, scale, 20)

	if tournamentData.Endowment > 0 {
		y = drawCenteredInfoLine(img, "DOTATION TOTALE", fmt.Sprintf("%d €", tournamentData.Endowment/100), y, scale)
		y += scaledSpacing(24, scale, 16)
	}

	dateStr := formatDates(tournamentData.StartDate, tournamentData.EndDate)
	dateLabel := "DATE"
	if tournamentData.StartDate != tournamentData.EndDate {
		dateLabel = "DATES"
	}
	y = drawCenteredInfoLine(img, dateLabel, dateStr, y, scale)
	y += scaledSpacing(24, scale, 16)

	clubName := wrapText(tournamentData.Club, 38, scale)
	y = drawCenteredInfoLine(img, "CLUB ORGANISATEUR", clubName, y, scale)
	y += scaledSpacing(24, scale, 16)

	address := wrapText(tournamentData.Address, 38, scale)
	y = drawCenteredInfoLine(img, "LIEU", address, y, scale)
	y += scaledSpacing(24, scale, 16)

	footerY := y + scaledSpacing(16, scale, 12)

	footerLabelSize := scaledFontSize(20, scale, 14)
	footerY = drawCenteredText(img, "RÈGLEMENT", footerY, ColorDarkGray, footerLabelSize, boldFont)
	footerY += scaledSpacing(6, scale, 5)

	urlText := wrapURL(tournamentData.TournamentURL, 38, scale)
	urlSize := scaledFontSize(24, scale, 16)
	finalY := drawCenteredText(img, urlText, footerY, ColorGradientStart, urlSize, boldFont)

	return finalY
}

func scaledFontSize(base float64, scale float64, min float64) float64 {
	size := base * scale
	if size < min {
		return min
	}
	return size
}

func scaledSpacing(base int, scale float64, min int) int {
	value := int(math.Round(float64(base) * scale))
	if value < min {
		return min
	}
	return value
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

// wrapText wraps text to fit within a certain character width
func wrapText(text string, maxWidth int, scale float64) string {
	if maxWidth <= 0 {
		return text
	}

	if scale < 1.0 {
		adjusted := int(math.Round(float64(maxWidth) / scale))
		if adjusted > maxWidth {
			maxWidth = adjusted
		}
	}

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

// wrapURL ensures URLs can break across multiple lines without being cut off
func wrapURL(url string, maxWidth int, scale float64) string {
	if maxWidth <= 0 {
		return url
	}

	if scale < 1.0 {
		adjusted := int(math.Round(float64(maxWidth) / scale))
		if adjusted > maxWidth {
			maxWidth = adjusted
		}
	}

	if len(url) <= maxWidth {
		return url
	}

	var lines []string
	remaining := url

	for len(remaining) > maxWidth {
		segment := remaining
		cut := maxWidth

		if cut < len(remaining) {
			if slash := strings.LastIndex(remaining[:cut], "/"); slash > 0 {
				cut = slash + 1
			}
		}

		segment = remaining[:cut]
		lines = append(lines, segment)
		remaining = remaining[cut:]
	}

	if len(remaining) > 0 {
		lines = append(lines, remaining)
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
func drawCenteredInfoLine(img *image.RGBA, label, text string, y int, scale float64) int {
	labelSize := scaledFontSize(20, scale, 14)
	textSize := scaledFontSize(32, scale, 18)
	offset := scaledSpacing(7, scale, 4)

	y = drawCenteredText(img, label, y, ColorDarkGray, labelSize, boldFont)
	endY := drawCenteredText(img, text, y+offset, ColorBlack, textSize, regularFont)
	return endY
}

// drawCenteredBadge draws a centered colored badge with text
func drawCenteredBadge(img *image.RGBA, text string, y int, bgColor color.RGBA, scale float64) int {
	// Get text dimensions
	faceSize := scaledFontSize(26, scale, 16)
	face, _ := opentype.NewFace(boldFont, &opentype.FaceOptions{
		Size:    faceSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	defer face.Close()

	// Measure text width
	drawer := &font.Drawer{Face: face}
	textWidth := drawer.MeasureString(text).Ceil()

	// Calculate centered position
	padding := scaledSpacing(18, scale, 10)
	badgeWidth := textWidth + (padding * 2)
	badgeHeight := scaledSpacing(42, scale, 24)
	x := (ImageWidth - badgeWidth) / 2

	// Fill badge background
	for row := y - 32; row < y-32+badgeHeight; row++ {
		for col := x; col < x+badgeWidth && col < ImageWidth; col++ {
			img.Set(col, row, bgColor)
		}
	}

	// Draw white text on badge (centered)
	return drawCenteredText(img, text, y, ColorWhite, faceSize, boldFont)
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
