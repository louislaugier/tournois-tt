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

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Brand colors
var (
	// Primary color - Tournois-TT cyan/turquoise
	ColorPrimary = color.RGBA{31, 186, 214, 255} // #1FBAD6
	// Secondary color - Purple
	ColorSecondary = color.RGBA{155, 89, 182, 255} // #9B59B6
	// Dark background
	ColorDark = color.RGBA{36, 39, 48, 255} // #242730
	// White text
	ColorWhite = color.RGBA{255, 255, 255, 255}
	// Light gray text
	ColorGray = color.RGBA{200, 200, 200, 255}
)

const (
	// Image dimensions optimized for Instagram
	ImageWidth  = 1080
	ImageHeight = 1080

	// Padding and margins
	Padding        = 60
	LineSpacing    = 35
	SectionSpacing = 50
)

// GenerateTournamentImage creates a tournament announcement image with brand identity
func GenerateTournamentImage(tournamentData TournamentImage) (string, error) {
	// Create a new image
	img := image.NewRGBA(image.Rect(0, 0, ImageWidth, ImageHeight))

	// Fill background with dark color
	draw.Draw(img, img.Bounds(), &image.Uniform{ColorDark}, image.Point{}, draw.Src)

	// Add header bar with primary color
	headerHeight := 180
	headerRect := image.Rect(0, 0, ImageWidth, headerHeight)
	draw.Draw(img, headerRect, &image.Uniform{ColorPrimary}, image.Point{}, draw.Src)

	// Add accent bar with secondary color
	accentHeight := 10
	accentRect := image.Rect(0, headerHeight, ImageWidth, headerHeight+accentHeight)
	draw.Draw(img, accentRect, &image.Uniform{ColorSecondary}, image.Point{}, draw.Src)

	// Draw content
	y := Padding + 20

	// Header: "NOUVEAU TOURNOI"
	y = drawText(img, "NOUVEAU TOURNOI", Padding, y, ColorWhite, 2.0)

	// Site name with smaller font
	y += 30
	y = drawText(img, "tournois-tt.fr", Padding, y, ColorGray, 1.2)

	// Move to content area
	y = headerHeight + accentHeight + SectionSpacing

	// Tournament name (title)
	y = drawText(img, wrapText(tournamentData.Name, 35), Padding, y, ColorPrimary, 1.8)
	y += SectionSpacing

	// Tournament details
	details := []struct {
		label string
		value string
		color color.RGBA
	}{
		{"Type:", tournamentData.Type, ColorWhite},
		{"Club:", wrapText(tournamentData.Club, 40), ColorWhite},
	}

	// Add endowment if available
	if tournamentData.Endowment > 0 {
		details = append(details, struct {
			label string
			value string
			color color.RGBA
		}{"Dotation:", fmt.Sprintf("%d €", tournamentData.Endowment), ColorSecondary})
	}

	// Add dates
	dateStr := formatDates(tournamentData.StartDate, tournamentData.EndDate)
	details = append(details, struct {
		label string
		value string
		color color.RGBA
	}{"Date(s):", dateStr, ColorWhite})

	// Add address
	details = append(details, struct {
		label string
		value string
		color color.RGBA
	}{"Lieu:", wrapText(tournamentData.Address, 40), ColorWhite})

	// Draw details
	for _, detail := range details {
		y = drawText(img, detail.label, Padding, y, ColorGray, 1.2)
		y += LineSpacing
		y = drawText(img, detail.value, Padding, y, detail.color, 1.3)
		y += LineSpacing + 15
	}

	// Footer section
	y = ImageHeight - 220

	// Separator line
	lineY := y - 30
	for x := Padding; x < ImageWidth-Padding; x++ {
		img.Set(x, lineY, ColorPrimary)
		img.Set(x, lineY+1, ColorPrimary)
	}

	// Tournament URL
	y = drawText(img, "Plus d'informations:", Padding, y, ColorGray, 1.1)
	y += LineSpacing
	y = drawText(img, tournamentData.TournamentURL, Padding, y, ColorPrimary, 1.4)

	y += SectionSpacing

	// Call to action
	y = drawText(img, "🎾 Consultez le règlement et inscrivez-vous!", Padding, y, ColorWhite, 1.2)

	// Save image to temporary file
	tmpDir := os.TempDir()
	filename := fmt.Sprintf("tournament_%d_%d.png", tournamentData.TournamentID, time.Now().Unix())
	filepath := filepath.Join(tmpDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()

	// Encode as PNG
	if err := png.Encode(file, img); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	return filepath, nil
}

// drawText draws text on the image at the specified position
// Returns the new Y position after the text
func drawText(img *image.RGBA, text string, x, y int, col color.RGBA, scale float64) int {
	// Use basic font (we'll use basicfont.Face7x13 scaled)
	// For production, you'd want to use truetype fonts
	face := basicfont.Face7x13

	point := fixed.Point26_6{
		X: fixed.Int26_6(x * 64),
		Y: fixed.Int26_6(y * 64),
	}

	drawer := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{col},
		Face: face,
		Dot:  point,
	}

	// Handle multi-line text
	lines := strings.Split(text, "\n")
	lineHeight := int(float64(face.Metrics().Height.Ceil()) * scale)

	for i, line := range lines {
		if i > 0 {
			y += lineHeight
			drawer.Dot = fixed.Point26_6{
				X: fixed.Int26_6(x * 64),
				Y: fixed.Int26_6(y * 64),
			}
		}
		drawer.DrawString(line)
	}

	return y + lineHeight
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
				// Single word is too long, just add it
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

// formatDates formats start and end dates for display
func formatDates(startDate, endDate string) string {
	if startDate == "" {
		return "Date non disponible"
	}

	// Parse dates
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return startDate
	}

	// If no end date or same as start date, show single date
	if endDate == "" || endDate == startDate {
		return start.Format("02/01/2006")
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return start.Format("02/01/2006")
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

