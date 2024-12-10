package types

// Club represents a table tennis club
type Club struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

// Rules represents tournament rules
type Rules struct {
	URL string `json:"url"`
}

// Table represents a tournament table (category)
type Table struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Time        string `json:"time"`
	Fee         int    `json:"fee"`       // in cents
	Endowment   int    `json:"endowment"` // in cents
}

// Organization represents a tournament organization
type Organization struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	ID         int    `json:"id"`
	Type       string `json:"@type"`
}

// Response represents a tournament response
type Response struct {
	Accountant   string       `json:"accountant"`
	Date         string       `json:"date"`
	Review       int          `json:"review"`
	Description  string       `json:"description"`
	Organization Organization `json:"organization"`
	ID           int          `json:"id"`
	Type         string       `json:"@type"`
}

// Tournament represents a table tennis tournament
type Tournament struct {
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	Type         string        `json:"type"`
	StartDate    string        `json:"startDate"`
	EndDate      string        `json:"endDate"`
	Address      Address       `json:"address"`
	Club         Club          `json:"club"`
	Rules        *Rules        `json:"rules,omitempty"`
	Tables       []Table       `json:"tables"`
	Status       int           `json:"status"`
	Endowment    int           `json:"endowment"` // in cents
	Organization *Organization `json:"organization,omitempty"`
	Responses    []Response    `json:"responses,omitempty"`
}
