package image

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

    "golang.org/x/image/font"
    "golang.org/x/image/font/gofont/gobold"
    "golang.org/x/image/font/gofont/goregular"
    "golang.org/x/image/font/opentype"
    "golang.org/x/image/math/fixed"
)

// Color palette and sizes
var (
    ColorGradientStart = color.RGBA{31, 186, 214, 255}
    ColorGradientEnd   = color.RGBA{155, 89, 182, 255}
    ColorWhite         = color.RGBA{255, 255, 255, 255}
    ColorLightGray     = color.RGBA{248, 249, 250, 255}
    ColorDarkGray      = color.RGBA{73, 80, 87, 255}
    ColorBlack         = color.RGBA{33, 37, 41, 255}
    ColorCardBg        = color.RGBA{255, 255, 255, 255}
    ColorBgStart       = color.RGBA{200, 230, 255, 255}  // More saturated blue
    ColorBgEnd         = color.RGBA{220, 200, 255, 255}  // More saturated purple
    ColorAccent        = color.RGBA{31, 186, 214, 255}
    ColorShadow        = color.RGBA{0, 0, 0, 20}         // Subtle shadow
)

const (
    ImageWidth        = 1080
    ImageHeight       = 1080
    StoryWidth        = 1080
    StoryHeight       = 1920
    gradientBarHeight = 12
    cardPadding       = 60
    cardMargin        = 80
)

var (
    regularFont *opentype.Font
    boldFont    *opentype.Font
)

func init() {
    var err error
    regularFont, err = opentype.Parse(goregular.TTF)
    if err != nil { panic(err) }
    boldFont, err = opentype.Parse(gobold.TTF)
    if err != nil { panic(err) }
}

// GenerateTournamentImage generates the feed image (1080x1080)
func GenerateTournamentImage(tournamentData TournamentImage) (string, error) {
    return generateImage(tournamentData, "feed")
}

// GenerateTournamentStoryImage generates the story image (1080x1920)
func GenerateTournamentStoryImage(tournamentData TournamentImage) (string, error) {
    return generateImage(tournamentData, "story")
}

func generateImage(tournamentData TournamentImage, imageType string) (string, error) {
    var img *image.RGBA
    var filename string
    
    if imageType == "story" {
        img = createStoryImage(tournamentData)
        timestamp := time.Now().Format("20060102-150405")
        filename = fmt.Sprintf("tournament_%d_%s_story.png", tournamentData.TournamentID, timestamp)
    } else {
        // Feed image with auto-scaling
        scales := []float64{1.0, 0.95, 0.9, 0.85, 0.8, 0.75}
        maxContentHeight := ImageHeight - (cardMargin * 2)
        var finalImg *image.RGBA
        finalY := math.MaxInt32
        bestOverflow := math.MaxInt32
        bestBottom := math.MaxInt32
        for idx, scale := range scales {
            testImg := createFeedImage()
            bottom, contentHeight := renderFeedContent(testImg, tournamentData, scale)
            overflow := contentHeight - maxContentHeight
            if overflow <= 0 { finalImg = testImg; finalY = bottom; break }
            if overflow < bestOverflow || (overflow == bestOverflow && bottom < bestBottom) {
                finalImg = testImg; finalY = bottom; bestOverflow = overflow; bestBottom = bottom
            }
            if idx == len(scales)-1 && finalImg == nil { finalImg = testImg; finalY = bottom }
        }
        if finalImg == nil { return "", fmt.Errorf("failed to render feed image") }
        img = finalImg
        _ = finalY
        timestamp := time.Now().Format("20060102-150405")
        filename = fmt.Sprintf("tournament_%d_%s.png", tournamentData.TournamentID, timestamp)
    }
    
    imagesDir := "./instagram-images"
    if err := os.MkdirAll(imagesDir, 0755); err != nil { 
        return "", fmt.Errorf("failed to create images directory: %w", err) 
    }
    filePath := filepath.Join(imagesDir, filename)
    file, err := os.Create(filePath)
    if err != nil { return "", fmt.Errorf("failed to create image file: %w", err) }
    defer file.Close()
    if err := png.Encode(file, img); err != nil { 
        return "", fmt.Errorf("failed to encode image: %w", err) 
    }
    return filePath, nil
}

// createFeedImage creates the base feed image (1080x1080) with gradient background
// Card dimensions will be set when rendering content
func createFeedImage() *image.RGBA {
    img := image.NewRGBA(image.Rect(0, 0, ImageWidth, ImageHeight))
    // Draw gradient background
    drawDiagonalGradient(img, 0, 0, ImageWidth, ImageHeight, ColorBgStart, ColorBgEnd)
    // Card will be drawn after measuring content
    return img
}

// createStoryImage creates the story image (1080x1920) with gradient background
func createStoryImage(tournamentData TournamentImage) *image.RGBA {
    img := image.NewRGBA(image.Rect(0, 0, StoryWidth, StoryHeight))
    // Draw gradient background
    drawDiagonalGradient(img, 0, 0, StoryWidth, StoryHeight, ColorBgStart, ColorBgEnd)
    
    // Calculate dynamic card dimensions based on content
    cardMarginStory := 40
    cardWidth := StoryWidth - cardMarginStory*2
    
    // Measure content height
    contentHeight := measureStoryContentHeight(tournamentData, cardWidth)
    
    // Add padding for card
    verticalPadding := 80
    cardHeight := contentHeight + (verticalPadding * 2)
    
    // Ensure card doesn't exceed reasonable bounds
    maxCardHeight := StoryHeight - 200 // leave 100px margin top and bottom
    if cardHeight > maxCardHeight {
        cardHeight = maxCardHeight
    }
    
    // Center card vertically
    cardY := (StoryHeight - cardHeight) / 2
    
    // Draw card
    drawCardWithShadow(img, cardMarginStory, cardY, cardWidth, cardHeight)
    
    // Render content
    renderStoryContent(img, tournamentData, cardMarginStory, cardY, cardWidth, cardHeight)
    return img
}

// renderFeedContent renders tournament content on feed image
func renderFeedContent(img *image.RGBA, tournamentData TournamentImage, scale float64) (int, int) {
    contentHeight := measureFeedContentHeight(tournamentData, scale)
    if contentHeight <= 0 { contentHeight = ImageHeight }
    
    // Calculate dynamic card dimensions
    horizontalMargin := 80
    verticalPadding := 60
    cardHeight := contentHeight + (verticalPadding * 2)
    
    // Ensure card doesn't exceed bounds
    maxCardHeight := ImageHeight - (horizontalMargin * 2)
    if cardHeight > maxCardHeight {
        cardHeight = maxCardHeight
    }
    
    // Center card vertically
    cardY := (ImageHeight - cardHeight) / 2
    cardX := horizontalMargin
    cardWidth := ImageWidth - (horizontalMargin * 2)
    
    // Draw dynamic card
    drawCardWithShadow(img, cardX, cardY, cardWidth, cardHeight)
    
    // Calculate content starting position (centered within card)
    startY := cardY + (cardHeight-contentHeight)/2
    if startY < cardY + 20 { startY = cardY + 20 }
    
    y := startY
    
    // Header without (FFTT)
    headerSize := scaledFontSize(40, scale, 26)
    y = drawCenteredText(img, "Nouvelle homologation", y, ColorBlack, headerSize, boldFont)
    y += scaledSpacing(35, scale, 22)
    
    // Tournament name
    tournamentName := wrapText(tournamentData.Name, 32, scale)
    tournamentNameWithQuotes := fmt.Sprintf("\"%s\"", tournamentName)
    nameSize := scaledFontSize(42, scale, 26)
    y = drawCenteredText(img, tournamentNameWithQuotes, y, ColorBlack, nameSize, boldFont)
    y += scaledSpacing(30, scale, 20)
    
    // Tournament type badge
    mappedType := utils.MapTournamentType(tournamentData.Type)
    y = drawCenteredBadge(img, mappedType, y, ColorGradientStart, scale)
    y += scaledSpacing(26, scale, 18)
    
    // Endowment
    if tournamentData.Endowment > 0 {
        y = drawCenteredInfoLine(img, "DOTATION TOTALE", fmt.Sprintf("%d €", tournamentData.Endowment/100), y, scale)
        y += scaledSpacing(22, scale, 14)
    }
    
    // Dates
    dateStr := formatDates(tournamentData.StartDate, tournamentData.EndDate)
    dateLabel := "DATE"
    if tournamentData.StartDate != tournamentData.EndDate { dateLabel = "DATES" }
    y = drawCenteredInfoLine(img, dateLabel, dateStr, y, scale)
    y += scaledSpacing(22, scale, 14)
    
    // Club
    clubName := wrapText(tournamentData.Club, 38, scale)
    y = drawCenteredInfoLine(img, "CLUB ORGANISATEUR", clubName, y, scale)
    y += scaledSpacing(22, scale, 14)
    
    // Address
    address := wrapText(tournamentData.Address, 38, scale)
    y = drawCenteredInfoLine(img, "LIEU", address, y, scale)
    y += scaledSpacing(28, scale, 16)
    
    // Footer
    footerLabelSize := scaledFontSize(18, scale, 12)
    y = drawCenteredText(img, "RÈGLEMENT", y, ColorDarkGray, footerLabelSize, boldFont)
    y += scaledSpacing(6, scale, 4)
    urlText := wrapURL(tournamentData.TournamentURL, 38, scale)
    urlSize := scaledFontSize(22, scale, 15)
    finalY := drawCenteredText(img, urlText, y, ColorGradientStart, urlSize, boldFont)
    
    return finalY, contentHeight
}

// renderStoryContent renders tournament content on story image
func renderStoryContent(img *image.RGBA, tournamentData TournamentImage, cardX, cardY, cardWidth, cardHeight int) {
    // Measure total content height first
    contentHeight := measureStoryContentHeight(tournamentData, cardWidth)
    
    // Center content vertically within card
    startY := cardY + (cardHeight-contentHeight)/2
    if startY < cardY+20 {
        startY = cardY + 20 // minimum top padding
    }
    
    y := startY
    
    // Header
    headerSize := 48.0
    y = drawCenteredTextInWidth(img, "Nouvelle homologation", y, ColorBlack, headerSize, boldFont, cardWidth, cardX)
    y += 40
    
    // Tournament name
    tournamentName := wrapText(tournamentData.Name, 28, 1.0)
    tournamentNameWithQuotes := fmt.Sprintf("\"%s\"", tournamentName)
    nameSize := 46.0
    y = drawCenteredTextInWidth(img, tournamentNameWithQuotes, y, ColorBlack, nameSize, boldFont, cardWidth, cardX)
    y += 35
    
    // Tournament type badge
    mappedType := utils.MapTournamentType(tournamentData.Type)
    y = drawCenteredBadgeInWidth(img, mappedType, y, ColorGradientStart, cardWidth, cardX)
    y += 30
    
    // Info lines
    if tournamentData.Endowment > 0 {
        y = drawCenteredInfoLineInWidth(img, "DOTATION TOTALE", fmt.Sprintf("%d €", tournamentData.Endowment/100), y, cardWidth, cardX)
        y += 25
    }
    
    dateStr := formatDates(tournamentData.StartDate, tournamentData.EndDate)
    dateLabel := "DATE"
    if tournamentData.StartDate != tournamentData.EndDate { dateLabel = "DATES" }
    y = drawCenteredInfoLineInWidth(img, dateLabel, dateStr, y, cardWidth, cardX)
    y += 25
    
    clubName := wrapText(tournamentData.Club, 32, 1.0)
    y = drawCenteredInfoLineInWidth(img, "CLUB ORGANISATEUR", clubName, y, cardWidth, cardX)
    y += 25
    
    address := wrapText(tournamentData.Address, 32, 1.0)
    y = drawCenteredInfoLineInWidth(img, "LIEU", address, y, cardWidth, cardX)
    y += 40
    
    // Footer with actual URL (like feed post)
    footerLabelSize := 22.0
    y = drawCenteredTextInWidth(img, "RÈGLEMENT", y, ColorDarkGray, footerLabelSize, boldFont, cardWidth, cardX)
    y += 8
    urlText := wrapURL(tournamentData.TournamentURL, 28, 1.0)
    urlSize := 26.0
    drawCenteredTextInWidth(img, urlText, y, ColorGradientStart, urlSize, boldFont, cardWidth, cardX)
}

// measureStoryContentHeight calculates the total height of story content
func measureStoryContentHeight(tournamentData TournamentImage, cardWidth int) int {
    totalHeight := 0
    
    // Header
    headerSize := 48.0
    totalHeight += int(headerSize * 1.5) // approximate height with line height
    totalHeight += 40
    
    // Tournament name
    tournamentName := wrapText(tournamentData.Name, 28, 1.0)
    tournamentNameWithQuotes := fmt.Sprintf("\"%s\"", tournamentName)
    nameLines := len(strings.Split(tournamentNameWithQuotes, "\n"))
    nameSize := 46.0
    totalHeight += int(nameSize*1.5) * nameLines
    totalHeight += 35
    
    // Badge
    totalHeight += 50 // badge height
    totalHeight += 30
    
    // Info lines
    if tournamentData.Endowment > 0 {
        totalHeight += 60 // label + value
        totalHeight += 25
    }
    
    totalHeight += 60 // date
    totalHeight += 25
    
    clubName := wrapText(tournamentData.Club, 32, 1.0)
    clubLines := len(strings.Split(clubName, "\n"))
    totalHeight += 60 * clubLines
    totalHeight += 25
    
    address := wrapText(tournamentData.Address, 32, 1.0)
    addressLines := len(strings.Split(address, "\n"))
    totalHeight += 60 * addressLines
    totalHeight += 40
    
    // Footer (label + URL)
    totalHeight += 33 // label
    totalHeight += 8  // spacing
    urlText := wrapURL(tournamentData.TournamentURL, 28, 1.0)
    urlLines := len(strings.Split(urlText, "\n"))
    totalHeight += 39 * urlLines // URL
    
    return totalHeight
}

func measureFeedContentHeight(tournamentData TournamentImage, scale float64) int {
    totalHeight := 0
    headerSize := scaledFontSize(40, scale, 26)
    totalHeight += measureCenteredTextHeight("Nouvelle homologation", headerSize, boldFont)
    totalHeight += scaledSpacing(35, scale, 22)
    tournamentName := wrapText(tournamentData.Name, 32, scale)
    tournamentNameWithQuotes := fmt.Sprintf("\"%s\"", tournamentName)
    nameSize := scaledFontSize(42, scale, 26)
    totalHeight += measureCenteredTextHeight(tournamentNameWithQuotes, nameSize, boldFont)
    totalHeight += scaledSpacing(30, scale, 20)
    mappedType := utils.MapTournamentType(tournamentData.Type)
    totalHeight += measureCenteredBadgeHeight(mappedType, scale)
    totalHeight += scaledSpacing(26, scale, 18)
    if tournamentData.Endowment > 0 {
        totalHeight += measureCenteredInfoLineHeight("DOTATION TOTALE", fmt.Sprintf("%d €", tournamentData.Endowment/100), scale)
        totalHeight += scaledSpacing(22, scale, 14)
    }
    dateStr := formatDates(tournamentData.StartDate, tournamentData.EndDate)
    dateLabel := "DATE"
    if tournamentData.StartDate != tournamentData.EndDate { dateLabel = "DATES" }
    totalHeight += measureCenteredInfoLineHeight(dateLabel, dateStr, scale)
    totalHeight += scaledSpacing(22, scale, 14)
    clubName := wrapText(tournamentData.Club, 38, scale)
    totalHeight += measureCenteredInfoLineHeight("CLUB ORGANISATEUR", clubName, scale)
    totalHeight += scaledSpacing(22, scale, 14)
    address := wrapText(tournamentData.Address, 38, scale)
    totalHeight += measureCenteredInfoLineHeight("LIEU", address, scale)
    totalHeight += scaledSpacing(28, scale, 16)
    footerLabelSize := scaledFontSize(18, scale, 12)
    totalHeight += measureCenteredTextHeight("RÈGLEMENT", footerLabelSize, boldFont)
    totalHeight += scaledSpacing(6, scale, 4)
    urlText := wrapURL(tournamentData.TournamentURL, 38, scale)
    urlSize := scaledFontSize(22, scale, 15)
    totalHeight += measureCenteredTextHeight(urlText, urlSize, boldFont)
    return totalHeight
}

func measureCenteredInfoLineHeight(label, text string, scale float64) int {
    labelSize := scaledFontSize(20, scale, 14)
    textSize := scaledFontSize(32, scale, 18)
    offset := scaledSpacing(7, scale, 4)
    labelHeight := measureCenteredTextHeight(label, labelSize, boldFont)
    textHeight := measureCenteredTextHeight(text, textSize, regularFont)
    return labelHeight + offset + textHeight
}

func measureCenteredBadgeHeight(text string, scale float64) int {
    badgeHeight := scaledSpacing(42, scale, 24)
    return badgeHeight
}

func measureCenteredTextHeight(text string, size float64, ttfFont *opentype.Font) int {
    face, err := opentype.NewFace(ttfFont, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
    if err != nil { return 0 }
    defer face.Close()
    metrics := face.Metrics()
    ascent := metrics.Ascent.Ceil()
    descent := metrics.Descent.Ceil()
    interline := int(size * 0.4)
    lines := strings.Split(text, "\n")
    if len(lines) == 0 { return ascent + descent }
    total := ascent + descent
    if len(lines) > 1 { total += (len(lines)-1) * (ascent + descent + interline) }
    return total
}

func scaledFontSize(base float64, scale float64, min float64) float64 { size := base * scale; if size < min { return min }; return size }
func scaledSpacing(base int, scale float64, min int) int { value := int(math.Round(float64(base) * scale)); if value < min { return min }; return value }

// drawDiagonalGradient draws a diagonal gradient background
func drawDiagonalGradient(img *image.RGBA, x, y, width, height int, startColor, endColor color.RGBA) {
    for row := y; row < y+height && row < img.Bounds().Max.Y; row++ {
        for col := x; col < x+width && col < img.Bounds().Max.X; col++ {
            // Diagonal gradient from top-left to bottom-right
            totalDist := float64(width + height)
            currentDist := float64(col-x) + float64(row-y)
            ratio := currentDist / totalDist
            if ratio > 1.0 { ratio = 1.0 }
            r := uint8(float64(startColor.R)*(1-ratio) + float64(endColor.R)*ratio)
            g := uint8(float64(startColor.G)*(1-ratio) + float64(endColor.G)*ratio)
            b := uint8(float64(startColor.B)*(1-ratio) + float64(endColor.B)*ratio)
            img.Set(col, row, color.RGBA{r, g, b, 255})
        }
    }
}

// drawCardWithShadow draws a white card (no shadow for cleaner look)
func drawCardWithShadow(img *image.RGBA, x, y, width, height int) {
    // Draw white card
    for row := y; row < y+height && row < img.Bounds().Max.Y; row++ {
        for col := x; col < x+width && col < img.Bounds().Max.X; col++ {
            img.Set(col, row, ColorCardBg)
        }
    }
}

// drawCenteredTextInWidth draws centered text within a specific width
func drawCenteredTextInWidth(img *image.RGBA, text string, y int, col color.RGBA, size float64, ttfFont *opentype.Font, width, offsetX int) int {
    face, err := opentype.NewFace(ttfFont, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
    if err != nil { return y }
    defer face.Close()
    drawer := &font.Drawer{Face: face}
    lines := strings.Split(text, "\n")
    metrics := face.Metrics()
    ascent := metrics.Ascent.Ceil()
    descent := metrics.Descent.Ceil()
    interline := int(size * 0.4)
    baseline := y + ascent
    for i, line := range lines {
        if i > 0 { baseline += ascent + descent + interline }
        textWidth := drawer.MeasureString(line).Ceil()
        x := offsetX + (width-textWidth)/2
        drawer.Dst = img
        drawer.Src = &image.Uniform{col}
        drawer.Dot = fixed.Point26_6{X: fixed.I(x), Y: fixed.I(baseline)}
        drawer.DrawString(line)
    }
    totalHeight := ascent + descent
    if len(lines) > 1 { totalHeight += (len(lines)-1) * (ascent + descent + interline) }
    return y + totalHeight
}

// drawCenteredBadgeInWidth draws centered badge within a specific width
func drawCenteredBadgeInWidth(img *image.RGBA, text string, y int, bgColor color.RGBA, width, offsetX int) int {
    faceSize := 30.0
    face, _ := opentype.NewFace(boldFont, &opentype.FaceOptions{Size: faceSize, DPI: 72, Hinting: font.HintingFull})
    defer face.Close()
    drawer := &font.Drawer{Face: face}
    textWidth := drawer.MeasureString(text).Ceil()
    padding := 20
    badgeWidth := textWidth + (padding * 2)
    badgeHeight := 50
    x := offsetX + (width-badgeWidth)/2
    badgeTop := y
    for row := badgeTop; row < badgeTop+badgeHeight && row < img.Bounds().Max.Y; row++ {
        for col := x; col < x+badgeWidth && col < img.Bounds().Max.X; col++ {
            img.Set(col, row, bgColor)
        }
    }
    metrics := face.Metrics()
    ascent := metrics.Ascent.Ceil()
    descent := metrics.Descent.Ceil()
    textBaseline := badgeTop + ((badgeHeight - (ascent + descent)) / 2) + ascent
    drawer.Dst = img
    drawer.Src = &image.Uniform{ColorWhite}
    drawer.Dot = fixed.Point26_6{X: fixed.I(x + padding), Y: fixed.I(textBaseline)}
    drawer.DrawString(text)
    return badgeTop + badgeHeight
}

// drawCenteredInfoLineInWidth draws centered info line within a specific width
func drawCenteredInfoLineInWidth(img *image.RGBA, label, text string, y int, width, offsetX int) int {
    labelSize := 22.0
    textSize := 34.0
    offset := 8
    y = drawCenteredTextInWidth(img, label, y, ColorDarkGray, labelSize, boldFont, width, offsetX)
    endY := drawCenteredTextInWidth(img, text, y+offset, ColorBlack, textSize, regularFont, width, offsetX)
    return endY
}

func drawGradientBar(img *image.RGBA, y, height int) {
    for row := y; row < y+height; row++ {
        for col := 0; col < ImageWidth; col++ {
            ratio := float64(col) / float64(ImageWidth)
            r := uint8(float64(ColorGradientStart.R)*(1-ratio) + float64(ColorGradientEnd.R)*ratio)
            g := uint8(float64(ColorGradientStart.G)*(1-ratio) + float64(ColorGradientEnd.G)*ratio)
            b := uint8(float64(ColorGradientStart.B)*(1-ratio) + float64(ColorGradientEnd.B)*ratio)
            img.Set(col, row, color.RGBA{r, g, b, 255})
        }
    }
}

func drawTextWithFont(img *image.RGBA, text string, x, y int, col color.RGBA, size float64, ttfFont *opentype.Font) int {
    face, err := opentype.NewFace(ttfFont, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
    if err != nil { return y }
    defer face.Close()
    drawer := &font.Drawer{Dst: img, Src: &image.Uniform{col}, Face: face, Dot: fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}}
    lines := strings.Split(text, "\n")
    metrics := face.Metrics()
    lineHeight := metrics.Ascent.Ceil() + metrics.Descent.Ceil() + int(size*0.4)
    currentY := y
    for i, line := range lines {
        if i > 0 { currentY += lineHeight; drawer.Dot = fixed.Point26_6{X: fixed.I(x), Y: fixed.I(currentY)} }
        drawer.DrawString(line)
    }
    return currentY + lineHeight
}

func drawCenteredInfoLine(img *image.RGBA, label, text string, y int, scale float64) int {
    labelSize := scaledFontSize(20, scale, 14); textSize := scaledFontSize(32, scale, 18); offset := scaledSpacing(7, scale, 4)
    y = drawCenteredText(img, label, y, ColorDarkGray, labelSize, boldFont)
    endY := drawCenteredText(img, text, y+offset, ColorBlack, textSize, regularFont)
    return endY
}

func drawCenteredBadge(img *image.RGBA, text string, y int, bgColor color.RGBA, scale float64) int {
    faceSize := scaledFontSize(26, scale, 16)
    face, _ := opentype.NewFace(boldFont, &opentype.FaceOptions{Size: faceSize, DPI: 72, Hinting: font.HintingFull})
    defer face.Close()
    drawer := &font.Drawer{Face: face}
    textWidth := drawer.MeasureString(text).Ceil()
    padding := scaledSpacing(18, scale, 10)
    badgeWidth := textWidth + (padding * 2)
    badgeHeight := scaledSpacing(42, scale, 24)
    x := (ImageWidth - badgeWidth) / 2
    badgeTop := y
    for row := badgeTop; row < badgeTop+badgeHeight; row++ {
        for col := x; col < x+badgeWidth && col < ImageWidth; col++ { img.Set(col, row, bgColor) }
    }
    metrics := face.Metrics(); ascent := metrics.Ascent.Ceil(); descent := metrics.Descent.Ceil()
    textBaseline := badgeTop + ((badgeHeight - (ascent + descent)) / 2) + ascent
    drawer.Dst = img; drawer.Src = &image.Uniform{ColorWhite}; drawer.Dot = fixed.Point26_6{X: fixed.I(x + padding), Y: fixed.I(textBaseline)}; drawer.DrawString(text)
    return badgeTop + badgeHeight
}

func wrapText(text string, maxWidth int, scale float64) string {
    if maxWidth <= 0 { return text }
    if scale < 1.0 { adjusted := int(math.Round(float64(maxWidth) / scale)); if adjusted > maxWidth { maxWidth = adjusted } }
    if len(text) <= maxWidth { return text }
    words := strings.Fields(text); var lines []string; var currentLine string
    for _, word := range words {
        testLine := currentLine; if testLine != "" { testLine += " " }; testLine += word
        if len(testLine) > maxWidth { if currentLine != "" { lines = append(lines, currentLine); currentLine = word } else { lines = append(lines, word); currentLine = "" } } else { currentLine = testLine }
    }
    if currentLine != "" { lines = append(lines, currentLine) }
    return strings.Join(lines, "\n")
}

func wrapURL(url string, maxWidth int, scale float64) string {
    if maxWidth <= 0 { return url }
    if scale < 1.0 { adjusted := int(math.Round(float64(maxWidth) / scale)); if adjusted > maxWidth { maxWidth = adjusted } }
    if len(url) <= maxWidth { return url }
    var lines []string; remaining := url
    for len(remaining) > maxWidth {
        cut := maxWidth; if cut < len(remaining) { if slash := strings.LastIndex(remaining[:cut], "/"); slash > 0 { cut = slash + 1 } }
        segment := remaining[:cut]; lines = append(lines, segment); remaining = remaining[cut:]
    }
    if len(remaining) > 0 { lines = append(lines, remaining) }
    return strings.Join(lines, "\n")
}

func drawCenteredText(img *image.RGBA, text string, y int, col color.RGBA, size float64, ttfFont *opentype.Font) int {
    face, err := opentype.NewFace(ttfFont, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull}); if err != nil { return y }
    defer face.Close()
    drawer := &font.Drawer{Face: face}
    lines := strings.Split(text, "\n")
    metrics := face.Metrics(); ascent := metrics.Ascent.Ceil(); descent := metrics.Descent.Ceil(); interline := int(size * 0.4)
    baseline := y + ascent
    for i, line := range lines {
        if i > 0 { baseline += ascent + descent + interline }
        textWidth := drawer.MeasureString(line).Ceil(); x := (ImageWidth - textWidth) / 2
        drawer.Dst = img; drawer.Src = &image.Uniform{col}; drawer.Dot = fixed.Point26_6{X: fixed.I(x), Y: fixed.I(baseline)}; drawer.DrawString(line)
    }
    totalHeight := ascent + descent; if len(lines) > 1 { totalHeight += (len(lines)-1) * (ascent + descent + interline) }
    return y + totalHeight
}

func CleanupImage(filepath string) error { return os.Remove(filepath) }

func formatDates(startDate, endDate string) string {
    if startDate == "" { return "Date non disponible" }
    var start, end time.Time; var err error
    start, err = time.Parse(time.RFC3339, startDate)
    if err != nil { if start, err = time.Parse("2006-01-02T15:04:05", startDate); err != nil { if start, err = time.Parse("2006-01-02", startDate); err != nil { return startDate } } }
    if endDate == "" || endDate == startDate { return start.Format("02/01/2006") }
    end, err = time.Parse(time.RFC3339, endDate)
    if err != nil { if end, err = time.Parse("2006-01-02T15:04:05", endDate); err != nil { if end, err = time.Parse("2006-01-02", endDate); err != nil { return start.Format("02/01/2006") } } }
    if start.Month() == end.Month() && start.Year() == end.Year() { return fmt.Sprintf("%d-%d %s %d", start.Day(), end.Day(), monthName(start.Month()), start.Year()) }
    return fmt.Sprintf("%s - %s", start.Format("02/01/2006"), end.Format("02/01/2006"))
}

func monthName(m time.Month) string { months := []string{"janvier","février","mars","avril","mai","juin","juillet","août","septembre","octobre","novembre","décembre"}; return months[m-1] }


