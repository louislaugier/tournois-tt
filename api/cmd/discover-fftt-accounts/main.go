package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// FFTTAccount represents an Instagram account to track
type FFTTAccount struct {
	Username   string    `json:"username"`
	Type       string    `json:"type"` // "federation", "ligue", "comite", "club"
	Region     string    `json:"region,omitempty"`
	Followed   bool      `json:"followed"`
	FollowedAt time.Time `json:"followed_at,omitempty"`
	Notes      string    `json:"notes,omitempty"`
}

// AccountList manages a list of FFTT accounts
type AccountList struct {
	Accounts    []FFTTAccount `json:"accounts"`
	LastUpdated time.Time     `json:"last_updated"`
}

const accountsFile = "./fftt-instagram-accounts.json"

func main() {
	// Initialize or load the accounts list
	accountList := loadOrCreateAccountList()

	// Add some known FFTT accounts (expand this list with real accounts)
	knownAccounts := []FFTTAccount{
		{
			Username: "fftt_officiel",
			Type:     "federation",
			Region:   "National",
			Followed: false,
			Notes:    "Official FFTT account",
		},
		// Add more accounts here - examples:
		// {
		//     Username: "ligue_idf_tt",
		//     Type: "ligue",
		//     Region: "Île-de-France",
		//     Followed: false,
		// },
		// {
		//     Username: "comite_75_tt",
		//     Type: "comite",
		//     Region: "Paris",
		//     Followed: false,
		// },
	}

	// Merge known accounts with existing list
	for _, known := range knownAccounts {
		exists := false
		for i, existing := range accountList.Accounts {
			if existing.Username == known.Username {
				// Update existing account metadata but preserve followed status
				accountList.Accounts[i].Type = known.Type
				accountList.Accounts[i].Region = known.Region
				accountList.Accounts[i].Notes = known.Notes
				exists = true
				break
			}
		}
		if !exists {
			accountList.Accounts = append(accountList.Accounts, known)
		}
	}

	accountList.LastUpdated = time.Now()

	// Save the updated list
	if err := saveAccountList(accountList); err != nil {
		log.Fatalf("Failed to save account list: %v", err)
	}

	// Print summary
	printSummary(accountList)
}

func loadOrCreateAccountList() *AccountList {
	data, err := os.ReadFile(accountsFile)
	if err != nil {
		// File doesn't exist, create new list
		return &AccountList{
			Accounts:    []FFTTAccount{},
			LastUpdated: time.Now(),
		}
	}

	var list AccountList
	if err := json.Unmarshal(data, &list); err != nil {
		log.Printf("Warning: Failed to parse accounts file, starting fresh: %v", err)
		return &AccountList{
			Accounts:    []FFTTAccount{},
			LastUpdated: time.Now(),
		}
	}

	return &list
}

func saveAccountList(list *AccountList) error {
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal account list: %w", err)
	}

	if err := os.WriteFile(accountsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write account list: %w", err)
	}

	log.Printf("✅ Saved account list to %s", accountsFile)
	return nil
}

func printSummary(list *AccountList) {
	fmt.Println("\n📊 FFTT Instagram Accounts Summary")
	fmt.Println("=====================================")
	fmt.Printf("Total accounts: %d\n", len(list.Accounts))
	fmt.Printf("Last updated: %s\n\n", list.LastUpdated.Format("2006-01-02 15:04:05"))

	byType := make(map[string]int)
	followed := 0
	notFollowed := 0

	for _, account := range list.Accounts {
		byType[account.Type]++
		if account.Followed {
			followed++
		} else {
			notFollowed++
		}
	}

	fmt.Println("By Type:")
	for accType, count := range byType {
		fmt.Printf("  - %s: %d\n", accType, count)
	}

	fmt.Printf("\nFollowed: %d\n", followed)
	fmt.Printf("Not followed: %d\n", notFollowed)

	fmt.Println("\n📝 Accounts to follow:")
	fmt.Println("=====================================")
	for _, account := range list.Accounts {
		if !account.Followed {
			fmt.Printf("- @%s (%s", account.Username, account.Type)
			if account.Region != "" {
				fmt.Printf(" - %s", account.Region)
			}
			fmt.Println(")")
		}
	}

	fmt.Println("\n💡 Next steps:")
	fmt.Println("1. Update the 'knownAccounts' list in this tool with more FFTT accounts")
	fmt.Println("2. The bot will automatically follow these accounts during daytime hours")
	fmt.Println("3. Run this tool periodically to add new accounts")
}
