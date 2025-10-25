package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"tournois-tt/api/pkg/memes"
)

func main() {
	// Command line flags
	listFlag := flag.Bool("list", false, "List all available meme templates")
	categoryFlag := flag.String("category", "", "Filter templates by category")
	idFlag := flag.String("id", "", "Generate meme by template ID")
	randomFlag := flag.Bool("random", false, "Generate a random meme")
	allFlag := flag.Bool("all", false, "Generate ALL memes")
	outputDir := flag.String("output", "./meme-output", "Output directory for memes")
	noBackgrounds := flag.Bool("no-bg", false, "Disable automatic backgrounds (plain text only)")

	flag.Parse()

	generator := memes.NewMemeGenerator(*outputDir)
	if *noBackgrounds {
		generator.SetBackgroundsEnabled(false)
	}
	templates := memes.GetAllTemplates()

	// List templates
	if *listFlag {
		if *categoryFlag != "" {
			templates = memes.GetTemplatesByCategory(*categoryFlag)
		}
		fmt.Printf("\n📋 Available Meme Templates (%d total)\n\n", len(templates))
		currentCategory := ""
		for _, t := range templates {
			if t.Category != currentCategory {
				currentCategory = t.Category
				fmt.Printf("\n🏷️  %s\n", currentCategory)
				fmt.Println("─────────────────────────────────────")
			}
			fmt.Printf("  [%s] (%s)\n", t.ID, t.Style)
			fmt.Printf("  → %s\n\n", t.Text)
		}
		return
	}

	// Generate by ID
	if *idFlag != "" {
		var selectedTemplate *memes.MemeTemplate
		for _, t := range templates {
			if t.ID == *idFlag {
				selectedTemplate = &t
				break
			}
		}
		if selectedTemplate == nil {
			log.Fatalf("❌ Template ID '%s' not found. Use -list to see available templates.", *idFlag)
		}

		log.Printf("🎬 Creating meme: %s (%s style)", selectedTemplate.ID, selectedTemplate.Style)
		outputPath, err := generator.GenerateMemeWithTemplate(
			*selectedTemplate,
			fmt.Sprintf("meme_%s.mp4", selectedTemplate.ID),
		)
		if err != nil {
			log.Fatalf("❌ Failed to create meme: %v", err)
		}
		fmt.Printf("✅ Meme created: %s\n", outputPath)
		return
	}

	// Generate random meme
	if *randomFlag {
		rand.Seed(time.Now().UnixNano())
		randomTemplate := templates[rand.Intn(len(templates))]

		log.Printf("🎲 Creating random meme: %s (%s style)", randomTemplate.ID, randomTemplate.Style)
		outputPath, err := generator.GenerateMemeWithTemplate(
			randomTemplate,
			fmt.Sprintf("meme_%s.mp4", randomTemplate.ID),
		)
		if err != nil {
			log.Fatalf("❌ Failed to create meme: %v", err)
		}
		fmt.Printf("✅ Random meme created: %s\n", outputPath)
		fmt.Printf("📝 Template: %s (%s category, %s style)\n", randomTemplate.ID, randomTemplate.Category, randomTemplate.Style)
		return
	}

	// Generate ALL memes
	if *allFlag {
		log.Printf("🎬 Creating ALL %d memes... This will take a while!", len(templates))
		successCount := 0
		failCount := 0

		for i, template := range templates {
			log.Printf("[%d/%d] Creating: %s (%s style)", i+1, len(templates), template.ID, template.Style)
			_, err := generator.GenerateMemeWithTemplate(
				template,
				fmt.Sprintf("meme_%s.mp4", template.ID),
			)
			if err != nil {
				log.Printf("  ⚠️  Failed: %v", err)
				failCount++
			} else {
				successCount++
			}
		}

		fmt.Printf("\n🎉 Batch complete!\n")
		fmt.Printf("  ✅ Success: %d\n", successCount)
		fmt.Printf("  ❌ Failed: %d\n", failCount)
		fmt.Printf("  📁 Output: %s\n", *outputDir)
		return
	}

	// No flags specified - show help
	fmt.Println("🏓 Tournois TT - Meme Generator")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  List templates:      -list [-category <category>]")
	fmt.Println("  Generate by ID:      -id <template_id>")
	fmt.Println("  Generate random:     -random")
	fmt.Println("  Generate all:        -all")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -output <dir>    Output directory (default: ./meme-output)")
	fmt.Println("  -no-bg           Disable backgrounds (plain text only)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/meme-generator/main.go -list")
	fmt.Println("  go run cmd/meme-generator/main.go -list -category fftt")
	fmt.Println("  go run cmd/meme-generator/main.go -id gratte_10_9")
	fmt.Println("  go run cmd/meme-generator/main.go -random")
	fmt.Println("  go run cmd/meme-generator/main.go -all")
	os.Exit(0)
}
