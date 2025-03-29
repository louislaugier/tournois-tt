package scripts

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// Create a custom logger without timestamps
var logger = log.New(os.Stdout, "", 0)

// Map of departments to regions
var departmentToRegion = map[string]string{
	// Auvergne-Rhône-Alpes
	"01": "Auvergne-Rhône-Alpes", "03": "Auvergne-Rhône-Alpes", "07": "Auvergne-Rhône-Alpes",
	"15": "Auvergne-Rhône-Alpes", "26": "Auvergne-Rhône-Alpes", "38": "Auvergne-Rhône-Alpes",
	"42": "Auvergne-Rhône-Alpes", "43": "Auvergne-Rhône-Alpes", "63": "Auvergne-Rhône-Alpes",
	"69": "Auvergne-Rhône-Alpes", "73": "Auvergne-Rhône-Alpes", "74": "Auvergne-Rhône-Alpes",
	// Bourgogne-Franche-Comté
	"21": "Bourgogne-Franche-Comté", "25": "Bourgogne-Franche-Comté", "39": "Bourgogne-Franche-Comté",
	"58": "Bourgogne-Franche-Comté", "70": "Bourgogne-Franche-Comté", "71": "Bourgogne-Franche-Comté",
	"89": "Bourgogne-Franche-Comté", "90": "Bourgogne-Franche-Comté",
	// Bretagne
	"22": "Bretagne", "29": "Bretagne", "35": "Bretagne", "56": "Bretagne",
	// Centre-Val de Loire
	"18": "Centre-Val de Loire", "28": "Centre-Val de Loire", "36": "Centre-Val de Loire",
	"37": "Centre-Val de Loire", "41": "Centre-Val de Loire", "45": "Centre-Val de Loire",
	// Corse
	"2A": "Corse", "2B": "Corse",
	// Grand Est
	"08": "Grand Est", "10": "Grand Est", "51": "Grand Est", "52": "Grand Est",
	"54": "Grand Est", "55": "Grand Est", "57": "Grand Est", "67": "Grand Est",
	"68": "Grand Est", "88": "Grand Est",
	// Hauts-de-France
	"02": "Hauts-de-France", "59": "Hauts-de-France", "60": "Hauts-de-France",
	"62": "Hauts-de-France", "80": "Hauts-de-France",
	// Île-de-France
	"75": "Île-de-France", "77": "Île-de-France", "78": "Île-de-France",
	"91": "Île-de-France", "92": "Île-de-France", "93": "Île-de-France",
	"94": "Île-de-France", "95": "Île-de-France",
	// Normandie
	"14": "Normandie", "27": "Normandie", "50": "Normandie",
	"61": "Normandie", "76": "Normandie",
	// Nouvelle-Aquitaine
	"16": "Nouvelle-Aquitaine", "17": "Nouvelle-Aquitaine", "19": "Nouvelle-Aquitaine",
	"23": "Nouvelle-Aquitaine", "24": "Nouvelle-Aquitaine", "33": "Nouvelle-Aquitaine",
	"40": "Nouvelle-Aquitaine", "47": "Nouvelle-Aquitaine", "64": "Nouvelle-Aquitaine",
	"79": "Nouvelle-Aquitaine", "86": "Nouvelle-Aquitaine", "87": "Nouvelle-Aquitaine",
	// Occitanie
	"09": "Occitanie", "11": "Occitanie", "12": "Occitanie", "30": "Occitanie",
	"31": "Occitanie", "32": "Occitanie", "34": "Occitanie", "46": "Occitanie",
	"48": "Occitanie", "65": "Occitanie", "66": "Occitanie", "81": "Occitanie",
	"82": "Occitanie",
	// Pays de la Loire
	"44": "Pays de la Loire", "49": "Pays de la Loire", "53": "Pays de la Loire",
	"72": "Pays de la Loire", "85": "Pays de la Loire",
	// Provence-Alpes-Côte d'Azur
	"04": "Provence-Alpes-Côte d'Azur", "05": "Provence-Alpes-Côte d'Azur",
	"06": "Provence-Alpes-Côte d'Azur", "13": "Provence-Alpes-Côte d'Azur",
	"83": "Provence-Alpes-Côte d'Azur", "84": "Provence-Alpes-Côte d'Azur",
	// Overseas departments
	"971": "Guadeloupe", "972": "Martinique", "973": "Guyane",
	"974": "La Réunion", "976": "Mayotte",
}

type Club struct {
	XMLName xml.Name `xml:"club"`
	Email   string   `xml:"mailcor"`
	Number  string   `xml:"numero"`
	Name    string   `xml:"nom"`
}

type ClubResponse struct {
	XMLName xml.Name `xml:"liste"`
	Clubs   []Club   `xml:"club"`
}

type ErrorResponse struct {
	XMLName xml.Name `xml:"erreurs"`
	Message string   `xml:",chardata"`
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "iso-8859-1":
		return transform.NewReader(input, charmap.ISO8859_1.NewDecoder()), nil
	default:
		return nil, fmt.Errorf("unknown charset: %s", charset)
	}
}

func getClubDetails(client *http.Client, clubID string, maxRetries int) (*ClubResponse, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, 8s, etc.
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(backoff)
		}

		resp, err := client.Get(fmt.Sprintf("https://fftt.dafunker.com/v1/proxy/xml_club_detail.php?club=%s", clubID))
		if err != nil {
			lastErr = err
			continue
		}

		// Check for 503 error
		if resp.StatusCode == http.StatusServiceUnavailable {
			resp.Body.Close()
			lastErr = fmt.Errorf("service unavailable (503)")
			continue
		}

		// Create XML decoder with charset reader
		decoder := xml.NewDecoder(resp.Body)
		decoder.CharsetReader = charsetReader

		// Try to decode as error response first
		var errorResp ErrorResponse
		if err := decoder.Decode(&errorResp); err == nil && errorResp.Message != "" {
			resp.Body.Close()
			return nil, fmt.Errorf("API error: %s", errorResp.Message)
		}

		// Reset response body for club data decoding
		resp.Body.Close()
		resp, err = client.Get(fmt.Sprintf("https://fftt.dafunker.com/v1/proxy/xml_club_detail.php?club=%s", clubID))
		if err != nil {
			lastErr = err
			continue
		}

		// Create new decoder for club data
		decoder = xml.NewDecoder(resp.Body)
		decoder.CharsetReader = charsetReader

		// Try to decode as club response
		var clubData ClubResponse
		err = decoder.Decode(&clubData)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		return &clubData, nil
	}
	return nil, fmt.Errorf("failed after %d attempts, last error: %v", maxRetries, lastErr)
}

func LogClubEmailAddresses() {
	// Create HTTP client with timeout and skip TLS verification
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// French department numbers (01-95 + overseas departments)
	departments := make([]string, 0)

	// Add metropolitan departments (01-95)
	for i := 1; i <= 95; i++ {
		departments = append(departments, fmt.Sprintf("%02d", i))
	}

	// Add overseas departments
	departments = append(departments, "971", "972", "973", "974", "976")

	// Regular expression to extract club IDs from HTML
	clubIDRegex := regexp.MustCompile(`structures/by-number\?number_id=(\d+)`)

	for _, dept := range departments {
		logger.Printf("Processing department: %s", dept)
		// Create form data
		formData := url.Values{}
		formData.Set("plugins_controller", "structures")
		formData.Set("plugins_action", "plugin_maps_ajax")
		formData.Set("structures_department", dept)

		// Make request to get clubs list
		resp, err := client.PostForm("https://www.fftt.com/site/ajax1", formData)
		if err != nil {
			logger.Printf("Error fetching clubs for department %s: %v", dept, err)
			continue
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			logger.Printf("Error reading response for department %s: %v", dept, err)
			continue
		}

		// Extract club IDs from HTML response
		matches := clubIDRegex.FindAllStringSubmatch(string(body), -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			clubID := match[1]

			// Get club details with retries
			clubData, err := getClubDetails(client, clubID, 3)
			if err != nil {
				logger.Printf("Error fetching details for club %s: %v", clubID, err)
				continue
			}

			// Log club email if available
			if len(clubData.Clubs) > 0 && clubData.Clubs[0].Email != "" {
				region := departmentToRegion[dept]
				if region == "" {
					region = "Unknown"
				}
				logger.Printf("%s;%s;%s;", strings.TrimSpace(clubData.Clubs[0].Email), dept, region)
			}

			// Add a small delay between requests
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func LogCommitteeAndLeagueEmailAddresses() {
	// Create HTTP client with timeout and skip TLS verification
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Create form data for leagues
	formData := url.Values{}
	formData.Set("plugins_controller", "structures")
	formData.Set("plugins_action", "plugin_maps_ajax")
	formData.Set("categories_id", "Ligue")

	// Make request to get both leagues and committees list
	resp, err := client.PostForm("https://www.fftt.com/site/ajax1", formData)
	if err != nil {
		logger.Printf("Error fetching structures: %v", err)
		return
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logger.Printf("Error reading response: %v", err)
		return
	}

	htmlContent := string(body)

	// Find the start of leagues and committees sections
	leaguesStart := strings.Index(htmlContent, "Les ligues")
	committeesStart := strings.Index(htmlContent, "Les comités")

	if leaguesStart == -1 || committeesStart == -1 {
		logger.Printf("Error: Could not find sections in HTML")
		return
	}

	// Process leagues
	logger.Printf("Processing leagues")
	leagueSection := htmlContent[leaguesStart:committeesStart]
	processStructureSection(client, leagueSection)

	// Process committees
	logger.Printf("Processing committees")
	committeeSection := htmlContent[committeesStart:]
	processStructureSection(client, committeeSection)
}

func processStructureSection(client *http.Client, section string) {
	// Extract structure IDs
	structureIDRegex := regexp.MustCompile(`structures/by-number\?number_id=(\d+)`)
	matches := structureIDRegex.FindAllStringSubmatch(section, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		structureID := match[1]

		// Get structure details
		resp, err := client.Get(fmt.Sprintf("https://www.fftt.com/site/structures/by-number?number_id=%s", structureID))
		if err != nil {
			logger.Printf("Error fetching details for structure %s: %v", structureID, err)
			continue
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			logger.Printf("Error reading response for structure %s: %v", structureID, err)
			continue
		}

		bodyStr := string(body)

		// Extract department number from the address
		deptRegex := regexp.MustCompile(`(?i)(?:^|\D)((?:97[1-4]|97[6]|2[AB]|0[1-9]|[1-9][0-9]))(?:\D|$)`)
		deptMatch := deptRegex.FindStringSubmatch(bodyStr)

		var region, dept string
		if len(deptMatch) > 1 {
			dept = deptMatch[1]
			region = departmentToRegion[dept]
		}
		if region == "" {
			// Try to extract region name directly from the page
			regionRegex := regexp.MustCompile(`(?i)Ligue\s+(?:de\s+)?([^<>\n]+)`)
			regionMatch := regionRegex.FindStringSubmatch(bodyStr)
			if len(regionMatch) > 1 {
				region = strings.TrimSpace(regionMatch[1])
			} else {
				region = "Unknown"
			}
		}

		// Extract all email addresses using regex
		emailRegex := regexp.MustCompile(`Mail : ([^<\n]+)(?:<|$)`)
		emailMatches := emailRegex.FindAllStringSubmatch(bodyStr, -1)

		for _, emailMatch := range emailMatches {
			if len(emailMatch) > 1 {
				// Split by various separators and clean each email
				emailStr := strings.TrimSpace(emailMatch[1])
				emails := strings.FieldsFunc(emailStr, func(r rune) bool {
					return r == '|' || r == ';' || r == ',' || r == '\n' || r == '\r'
				})

				for _, email := range emails {
					if email = strings.TrimSpace(email); email != "" {
						if dept != "" {
							logger.Printf("%s;%s;%s;", email, dept, region)
						} else {
							logger.Printf("%s;;%s;", email, region)
						}
					}
				}
			}
		}

		// Add a small delay between requests
		time.Sleep(200 * time.Millisecond)
	}
}
