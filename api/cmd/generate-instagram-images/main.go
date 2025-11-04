package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strconv"
    "strings"

    "tournois-tt/api/pkg/cache"
    igimage "tournois-tt/api/pkg/image"
)

func main() {
    idsFlag := flag.String("ids", "", "Comma-separated list of tournament IDs")
    storyFlag := flag.Bool("story", true, "Generate story images (1080x1920) in addition to feed images")
    feedFlag := flag.Bool("feed", true, "Generate feed images (1080x1080)")
    flag.Parse()

    if *idsFlag == "" {
        log.Fatal("please provide tournament IDs via --ids")
    }

    ids, err := parseIDs(*idsFlag)
    if err != nil {
        log.Fatalf("invalid ids: %v", err)
    }

    tournaments, err := loadTournaments()
    if err != nil {
        log.Fatalf("failed to load tournaments: %v", err)
    }

    index := make(map[int]cache.TournamentCache)
    for _, t := range tournaments {
        index[t.ID] = t
    }

    for _, id := range ids {
        tournament, ok := index[id]
        if !ok {
            log.Printf("⚠️  Tournament %d not found in cache", id)
            continue
        }

        tournamentImage := convertToImage(tournament)
        
        // Generate feed image
        if *feedFlag {
            imagePath, err := igimage.GenerateTournamentImage(tournamentImage)
            if err != nil {
                log.Printf("❌ Failed to generate feed image for %d (%s): %v", id, tournament.Name, err)
            } else {
                fmt.Printf("✅ Generated feed image for %d (%s): %s\n", id, tournament.Name, imagePath)
            }
        }
        
        // Generate story image
        if *storyFlag {
            storyPath, err := igimage.GenerateTournamentStoryImage(tournamentImage)
            if err != nil {
                log.Printf("❌ Failed to generate story image for %d (%s): %v", id, tournament.Name, err)
            } else {
                fmt.Printf("✅ Generated story image for %d (%s): %s\n", id, tournament.Name, storyPath)
            }
        }
    }
}

func parseIDs(raw string) ([]int, error) {
    parts := strings.Split(raw, ",")
    ids := make([]int, 0, len(parts))
    for _, part := range parts {
        trimmed := strings.TrimSpace(part)
        if trimmed == "" {
            continue
        }
        id, err := strconv.Atoi(trimmed)
        if err != nil {
            return nil, err
        }
        ids = append(ids, id)
    }
    return ids, nil
}

func loadTournaments() ([]cache.TournamentCache, error) {
    possiblePaths := []string{
        "./api/cache/data.json",
        "./cache/data.json",
        "../cache/data.json",
        "../../cache/data.json",
    }

    var dataPath string
    for _, path := range possiblePaths {
        if _, err := os.Stat(path); err == nil {
            dataPath = path
            break
        }
    }

    if dataPath == "" {
        wd, _ := os.Getwd()
        current := wd
        for current != "/" && current != "." {
            candidate := filepath.Join(current, "api", "cache", "data.json")
            if _, err := os.Stat(candidate); err == nil {
                dataPath = candidate
                break
            }

            candidate = filepath.Join(current, "cache", "data.json")
            if _, err := os.Stat(candidate); err == nil {
                dataPath = candidate
                break
            }

            current = filepath.Dir(current)
        }
    }

    if dataPath == "" {
        return nil, fmt.Errorf("data.json not found")
    }

    payload, err := os.ReadFile(dataPath)
    if err != nil {
        return nil, err
    }

    var tournaments []cache.TournamentCache
    if err := json.Unmarshal(payload, &tournaments); err != nil {
        return nil, err
    }

    return tournaments, nil
}

func convertToImage(t cache.TournamentCache) igimage.TournamentImage {
    address := formatAddress(t.Address)

    rulesURL := ""
    if t.Rules != nil && t.Rules.URL != "" {
        rulesURL = t.Rules.URL
    }

    tournamentURL := fmt.Sprintf("https://tournois-tt.fr/%d", t.ID)

    clubName := t.Club.Name
    if t.Club.Identifier != "" {
        clubName = fmt.Sprintf("%s (%s)", t.Club.Name, t.Club.Identifier)
    }

    return igimage.TournamentImage{
        Name:          t.Name,
        Type:          t.Type,
        Club:          clubName,
        Endowment:     t.Endowment,
        StartDate:     t.StartDate,
        EndDate:       t.EndDate,
        Address:       address,
        RulesURL:      rulesURL,
        TournamentID:  t.ID,
        TournamentURL: tournamentURL,
    }
}

func formatAddress(addr cache.Address) string {
    parts := []string{}

    if addr.DisambiguatingDescription != "" {
        parts = append(parts, addr.DisambiguatingDescription)
    }

    if addr.StreetAddress != "" {
        parts = append(parts, addr.StreetAddress)
    }

    locality := strings.TrimSpace(addr.AddressLocality)
    if addr.PostalCode != "" && locality != "" {
        parts = append(parts, fmt.Sprintf("%s %s", addr.PostalCode, locality))
    } else if locality != "" {
        parts = append(parts, locality)
    }

    if len(parts) == 0 {
        return "Adresse non disponible"
    }

    return strings.Join(parts, ", ")
}

