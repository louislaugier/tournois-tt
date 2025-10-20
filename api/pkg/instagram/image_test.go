package instagram

import (
	"os"
	"testing"
)

func TestGenerateTournamentImage(t *testing.T) {
	// Test data
	tournament := TournamentImage{
		Name:          "Tournoi National de Paris",
		Type:          "Tournoi jeunes",
		Club:          "Paris TT Club",
		Endowment:     1500,
		StartDate:     "2025-11-15",
		EndDate:       "2025-11-16",
		Address:       "Gymnase Jean Jaurès, 123 Rue du Sport, 75001 Paris",
		RulesURL:      "https://example.com/rules.pdf",
		TournamentID:  12345,
		TournamentURL: "https://tournois-tt.fr/12345",
	}

	// Generate image
	imagePath, err := GenerateTournamentImage(tournament)
	if err != nil {
		t.Fatalf("Failed to generate image: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Fatalf("Image file was not created: %s", imagePath)
	}

	// Verify file size (should be > 0)
	fileInfo, err := os.Stat(imagePath)
	if err != nil {
		t.Fatalf("Failed to stat image file: %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Fatalf("Image file is empty")
	}

	t.Logf("Image generated successfully: %s (size: %d bytes)", imagePath, fileInfo.Size())

	// Cleanup
	if err := CleanupImage(imagePath); err != nil {
		t.Logf("Warning: Failed to cleanup image: %v", err)
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxWidth  int
		wantLines int
	}{
		{
			name:      "Short text",
			text:      "Short",
			maxWidth:  20,
			wantLines: 1,
		},
		{
			name:      "Long text wraps",
			text:      "This is a very long tournament name that should wrap",
			maxWidth:  20,
			wantLines: 3,
		},
		{
			name:      "Exact width",
			text:      "Exactly twenty chars",
			maxWidth:  20,
			wantLines: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapText(tt.text, tt.maxWidth)
			lines := 1
			for _, c := range result {
				if c == '\n' {
					lines++
				}
			}

			if lines != tt.wantLines {
				t.Errorf("wrapText() got %d lines, want %d lines\nInput: %q\nOutput: %q",
					lines, tt.wantLines, tt.text, result)
			}
		})
	}
}

func TestFormatDates(t *testing.T) {
	tests := []struct {
		name      string
		startDate string
		endDate   string
		want      string
	}{
		{
			name:      "Single day",
			startDate: "2025-11-15",
			endDate:   "2025-11-15",
			want:      "15/11/2025",
		},
		{
			name:      "Same month range",
			startDate: "2025-11-15",
			endDate:   "2025-11-16",
			want:      "15-16 novembre 2025",
		},
		{
			name:      "Different months",
			startDate: "2025-11-30",
			endDate:   "2025-12-01",
			want:      "30/11/2025 - 01/12/2025",
		},
		{
			name:      "No end date",
			startDate: "2025-11-15",
			endDate:   "",
			want:      "15/11/2025",
		},
		{
			name:      "No start date",
			startDate: "",
			endDate:   "2025-11-15",
			want:      "Date non disponible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDates(tt.startDate, tt.endDate)
			if got != tt.want {
				t.Errorf("formatDates() = %q, want %q", got, tt.want)
			}
		})
	}
}

