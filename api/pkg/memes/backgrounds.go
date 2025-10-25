package memes

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// BackgroundManager handles managing video backgrounds
type BackgroundManager struct {
	templatesDir string
	httpClient   *http.Client
}

// NewBackgroundManager creates a new background manager
func NewBackgroundManager(templatesDir string) *BackgroundManager {
	return &BackgroundManager{
		templatesDir: templatesDir,
		httpClient:   &http.Client{},
	}
}

// MemeTemplate represents a meme video template source
type MemeTemplateSource struct {
	Style       string
	Filename    string
	DirectURL   string // Direct download URL
	Description string
}

// GetMemeTemplateSources returns a curated list of meme video templates with direct URLs
// These are hosted on various CDNs and free to use
func GetMemeTemplateSources() []MemeTemplateSource {
	return []MemeTemplateSource{
		// Panic/Stress/Nervous
		{
			Style:       "panic",
			Filename:    "panic.mp4",
			DirectURL:   "https://media.giphy.com/media/LRZc4dV2kf1dtZOTlh/giphy.mp4",
			Description: "Sweating/Nervous reaction",
		},
		{
			Style:       "nervous",
			Filename:    "nervous.mp4",
			DirectURL:   "https://media.giphy.com/media/32mC2kXYWCsg0/giphy.mp4",
			Description: "Nervous/Anxious",
		},
		{
			Style:       "pressure",
			Filename:    "pressure.mp4",
			DirectURL:   "https://media.giphy.com/media/l0IylOPCNkiqOgMyA/giphy.mp4",
			Description: "Under pressure/stress",
		},

		// Celebration/Joy/Victory
		{
			Style:       "celebration",
			Filename:    "celebration.mp4",
			DirectURL:   "https://media.giphy.com/media/g9582DNuQppxC/giphy.mp4",
			Description: "Celebration/Success",
		},
		{
			Style:       "hype",
			Filename:    "hype.mp4",
			DirectURL:   "https://media.giphy.com/media/3ohzdIuqJoo8QdKlnW/giphy.mp4",
			Description: "Excited/Hyped",
		},
		{
			Style:       "comeback",
			Filename:    "comeback.mp4",
			DirectURL:   "https://media.giphy.com/media/YJ5OlVLZ2QNl6/giphy.mp4",
			Description: "Victory/Comeback",
		},
		{
			Style:       "legend",
			Filename:    "legend.mp4",
			DirectURL:   "https://media.giphy.com/media/xT0xeJpnrWC4XWblEk/giphy.mp4",
			Description: "Legendary moment",
		},

		// Frustration/Anger/Rage
		{
			Style:       "frustration",
			Filename:    "frustration.mp4",
			DirectURL:   "https://media.giphy.com/media/3o7btT1T9qpQZWhNlK/giphy.mp4",
			Description: "Frustrated/Disappointed",
		},
		{
			Style:       "anger",
			Filename:    "anger.mp4",
			DirectURL:   "https://media.giphy.com/media/yhLV2DGTLDRCw/giphy.mp4",
			Description: "Angry reaction",
		},
		{
			Style:       "rage",
			Filename:    "rage.mp4",
			DirectURL:   "https://media.giphy.com/media/3oEjHIKrinRbYbwvfO/giphy.mp4",
			Description: "Extreme anger/rage",
		},
		{
			Style:       "triggered",
			Filename:    "triggered.mp4",
			DirectURL:   "https://media.giphy.com/media/26uf4r3EldfX3bbKo/giphy.mp4",
			Description: "Triggered/Offended",
		},
		{
			Style:       "facepalm",
			Filename:    "facepalm.mp4",
			DirectURL:   "https://media.giphy.com/media/XsUtdIeJ0MWMo/giphy.mp4",
			Description: "Facepalm reaction",
		},

		// Confusion/Thinking
		{
			Style:       "confused",
			Filename:    "confused.mp4",
			DirectURL:   "https://media.giphy.com/media/WRQBXSCnEFJIuxktnw/giphy.mp4",
			Description: "Confused/Thinking",
		},
		{
			Style:       "realization",
			Filename:    "realization.mp4",
			DirectURL:   "https://media.giphy.com/media/3o7qDEq2bMbcbPRQ2c/giphy.mp4",
			Description: "Sudden realization",
		},
		{
			Style:       "existential",
			Filename:    "existential.mp4",
			DirectURL:   "https://media.giphy.com/media/3o7TKTDn976rzVgky4/giphy.mp4",
			Description: "Deep thinking/existential",
		},

		// Disappointment/Sadness
		{
			Style:       "disappointed",
			Filename:    "disappointed.mp4",
			DirectURL:   "https://media.giphy.com/media/3oEjHCWdU7F4hkcudy/giphy.mp4",
			Description: "Disappointed/Sad",
		},
		{
			Style:       "devastated",
			Filename:    "devastated.mp4",
			DirectURL:   "https://media.giphy.com/media/d2lcHJTG5Tscg/giphy.mp4",
			Description: "Shocked/Devastated",
		},
		{
			Style:       "pain",
			Filename:    "pain.mp4",
			DirectURL:   "https://media.giphy.com/media/3o7qDDNMYsLOYb2rjW/giphy.mp4",
			Description: "Pain/Suffering",
		},
		{
			Style:       "regret",
			Filename:    "regret.mp4",
			DirectURL:   "https://media.giphy.com/media/26xBwdIuRJiAIqHwA/giphy.mp4",
			Description: "Regret/Mistake",
		},
		{
			Style:       "disappointment",
			Filename:    "disappointment.mp4",
			DirectURL:   "https://media.giphy.com/media/ISOckXUybVfQ4/giphy.mp4",
			Description: "General disappointment",
		},

		// Waiting/Boredom
		{
			Style:       "waiting",
			Filename:    "waiting.mp4",
			DirectURL:   "https://media.giphy.com/media/tXL4FHPSnVJ0A/giphy.mp4",
			Description: "Waiting/Bored",
		},
		{
			Style:       "resigned",
			Filename:    "resigned.mp4",
			DirectURL:   "https://media.giphy.com/media/3oEjI5VtIhHvK37WYo/giphy.mp4",
			Description: "Resigned/Accepting defeat",
		},
		{
			Style:       "lost",
			Filename:    "lost.mp4",
			DirectURL:   "https://media.giphy.com/media/hEc4k5pN17GZq/giphy.mp4",
			Description: "Lost/Confused wandering",
		},

		// Awkward/Embarrassed
		{
			Style:       "awkward",
			Filename:    "awkward.mp4",
			DirectURL:   "https://media.giphy.com/media/kGCuRgmbnO9EI/giphy.mp4",
			Description: "Awkward silence",
		},
		{
			Style:       "embarrassed",
			Filename:    "embarrassed.mp4",
			DirectURL:   "https://media.giphy.com/media/13XFmJhNtbTSgM/giphy.mp4",
			Description: "Embarrassed/Cringe",
		},

		// Intense/Obsessive
		{
			Style:       "obsessive",
			Filename:    "obsessive.mp4",
			DirectURL:   "https://media.giphy.com/media/5wWf7H89PisM6An8UAU/giphy.mp4",
			Description: "Intense/Focused",
		},
		{
			Style:       "desperate",
			Filename:    "desperate.mp4",
			DirectURL:   "https://media.giphy.com/media/l2JhIUyUs8KDCCf3W/giphy.mp4",
			Description: "Desperate/Urgent",
		},

		// Cope/Relatable
		{
			Style:       "cope",
			Filename:    "cope.mp4",
			DirectURL:   "https://media.giphy.com/media/QMHoU66sBXqqLqYvGO/giphy.mp4",
			Description: "Coping/Dealing with it",
		},
		{
			Style:       "relatable",
			Filename:    "relatable.mp4",
			DirectURL:   "https://media.giphy.com/media/l0HlBO7eyXzSZkJri/giphy.mp4",
			Description: "Relatable situations",
		},
		{
			Style:       "procrastination",
			Filename:    "procrastination.mp4",
			DirectURL:   "https://media.giphy.com/media/26n6ziTEeDDbowBkQ/giphy.mp4",
			Description: "Lazy/Procrastinating",
		},
		{
			Style:       "transformation",
			Filename:    "transformation.mp4",
			DirectURL:   "https://media.giphy.com/media/3oz8xsaLEJ7G0zETCw/giphy.mp4",
			Description: "Change/Transformation",
		},

		// Generic fallback
		{
			Style:       "generic",
			Filename:    "generic.mp4",
			DirectURL:   "https://media.giphy.com/media/l0HlRnAWXxn0MhKLK/giphy.mp4",
			Description: "Generic reaction",
		},
	}
}

// StyleToTemplate maps meme styles to template filenames
var StyleToTemplate = map[string]string{
	// Panic/Stress
	"panic":    "panic.mp4",
	"nervous":  "nervous.mp4",
	"pressure": "pressure.mp4",

	// Frustration/Anger
	"frustration": "frustration.mp4",
	"anger":       "anger.mp4",
	"rage":        "rage.mp4",
	"triggered":   "triggered.mp4",
	"facepalm":    "facepalm.mp4",

	// Celebration/Joy
	"celebration":    "celebration.mp4",
	"hype":           "hype.mp4",
	"comeback":       "comeback.mp4",
	"transformation": "transformation.mp4",
	"legend":         "legend.mp4",

	// Waiting/Boredom
	"waiting":  "waiting.mp4",
	"resigned": "resigned.mp4",
	"lost":     "lost.mp4",

	// Confusion
	"confused":        "confused.mp4",
	"realization":     "realization.mp4",
	"existential":     "existential.mp4",
	"procrastination": "procrastination.mp4",

	// Sadness/Disappointment
	"disappointed":   "disappointed.mp4",
	"devastated":     "devastated.mp4",
	"pain":           "pain.mp4",
	"disappointment": "disappointment.mp4",
	"regret":         "regret.mp4",

	// Awkward/Embarrassed
	"awkward":     "awkward.mp4",
	"embarrassed": "embarrassed.mp4",

	// Obsessive/Intense
	"obsessive": "obsessive.mp4",
	"desperate": "desperate.mp4",

	// Relatable/Cope
	"relatable": "relatable.mp4",
	"cope":      "cope.mp4",
	"generic":   "generic.mp4",
}

// GetBackgroundForStyle returns the path to a video template for a given style
func (b *BackgroundManager) GetBackgroundForStyle(style string) (string, error) {
	// Ensure templates directory exists
	if err := os.MkdirAll(b.templatesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create templates dir: %w", err)
	}

	// Get template filename for this style
	templateName, ok := StyleToTemplate[style]
	if !ok {
		templateName = "generic.mp4"
	}

	templatePath := filepath.Join(b.templatesDir, templateName)

	// Check if template already exists (cached)
	if _, err := os.Stat(templatePath); err == nil {
		return templatePath, nil
	}

	// Template doesn't exist, try to download it
	fmt.Printf("📥 Downloading template: %s...\n", templateName)
	if err := b.downloadTemplate(style, templateName); err != nil {
		return "", fmt.Errorf("failed to download template: %w", err)
	}

	return templatePath, nil
}

// downloadTemplate downloads a meme template from the source list
func (b *BackgroundManager) downloadTemplate(style, filename string) error {
	// Find the source for this template
	sources := GetMemeTemplateSources()
	var sourceURL string

	for _, source := range sources {
		if source.Filename == filename {
			sourceURL = source.DirectURL
			break
		}
	}

	if sourceURL == "" {
		return fmt.Errorf("no source URL found for template: %s", filename)
	}

	// Download the video
	resp, err := b.httpClient.Get(sourceURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Save to file
	filepath := filepath.Join(b.templatesDir, filename)
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Printf("✅ Downloaded: %s\n", filename)
	return nil
}

// ListAvailableTemplates returns all available template videos
func (b *BackgroundManager) ListAvailableTemplates() ([]string, error) {
	files, err := os.ReadDir(b.templatesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var templates []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".mp4") {
			templates = append(templates, file.Name())
		}
	}
	return templates, nil
}
