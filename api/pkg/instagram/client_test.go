package instagram

import (
	"testing"
)

// TestClient_PostTournament_MockIntegration demonstrates how we would test with mocks
// Currently skipped because it would require modifying client.go to accept a custom HTTP client
// This prevents accidentally posting real Instagram posts during testing
func TestClient_PostTournament_MockIntegration(t *testing.T) {
	t.Skip("Skipping mock integration test - would require dependency injection for HTTP client")

	// To implement this properly, we would need to:
	// 1. Modify client.go to accept an http.Client via dependency injection
	// 2. Create a mock HTTP server that responds to Instagram API endpoints
	// 3. Inject the mock HTTP client into Instagram client
	// 4. Call PostTournament()
	// 5. Verify mock endpoints were called correctly
	// 6. Verify no real API calls were made
	//
	// For now, we rely on:
	// - Unit tests for image generation (safe, no API calls)
	// - Manual testing with real Instagram credentials
	// - INSTAGRAM_ENABLED=false in tests prevents accidental posts
}

// TestClient_TestConnection tests the connection test with disabled Instagram
func TestClient_TestConnection_Disabled(t *testing.T) {
	config := Config{
		AccessToken: "test-token",
		PageID:      "test-page",
		RecipientID: "test-recipient",
		Enabled:     false,
	}

	client := NewClient(config)
	err := client.TestConnection()

	if err == nil {
		t.Error("Expected error when Instagram is disabled, got nil")
	}

	expectedMsg := "Instagram integration is disabled"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

// TestClient_PostTournament_Disabled tests that posting is skipped when disabled
func TestClient_PostTournament_Disabled(t *testing.T) {
	config := Config{
		AccessToken: "test-token",
		PageID:      "test-page",
		RecipientID: "test-recipient",
		Enabled:     false,
	}

	client := NewClient(config)

	tournament := TournamentImage{
		Name:          "Test Tournament",
		Type:          "Tournoi jeunes",
		Club:          "Test Club",
		Endowment:     500,
		StartDate:     "2025-11-15",
		EndDate:       "2025-11-15",
		Address:       "Test Address",
		TournamentID:  999,
		TournamentURL: "https://tournois-tt.fr/999",
	}

	notification, err := client.PostTournament(tournament)

	if err == nil {
		t.Error("Expected error when Instagram is disabled, got nil")
	}

	if notification != nil {
		t.Errorf("Expected nil notification, got %+v", notification)
	}
}

// TestConfig_Validation tests configuration validation
func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		wantEnabled bool
	}{
		{
			name: "Fully configured",
			config: Config{
				AccessToken: "token",
				PageID:      "page",
				RecipientID: "recipient",
				Enabled:     true,
			},
			wantEnabled: true,
		},
		{
			name: "Disabled",
			config: Config{
				AccessToken: "token",
				PageID:      "page",
				RecipientID: "recipient",
				Enabled:     false,
			},
			wantEnabled: false,
		},
		{
			name: "Missing token",
			config: Config{
				AccessToken: "",
				PageID:      "page",
				RecipientID: "recipient",
				Enabled:     true,
			},
			wantEnabled: true, // Config doesn't validate, just stores
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.config)
			if client == nil {
				t.Fatal("Expected client to be created, got nil")
			}

			if client.config.Enabled != tt.wantEnabled {
				t.Errorf("Expected Enabled=%v, got %v", tt.wantEnabled, client.config.Enabled)
			}
		})
	}
}

// Benchmark tests for image generation
func BenchmarkGenerateTournamentImage(b *testing.B) {
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		imagePath, err := GenerateTournamentImage(tournament)
		if err != nil {
			b.Fatalf("Failed to generate image: %v", err)
		}
		CleanupImage(imagePath)
	}
}

// TestConvertTournamentToImageData tests the conversion logic (would be in cache_test.go)
// This is a conceptual test showing what SHOULD be tested
func TestTournamentImageConversion(t *testing.T) {
	t.Skip("This would test the convertTournamentToImageData function in cache.go")
	// In reality, this test should be in api/pkg/cache/cache_test.go
	// Testing the conversion from TournamentCache to TournamentImage
}
