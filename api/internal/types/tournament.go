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

// Tournament represents a table tennis tournament
type Tournament struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	StartDate string  `json:"startDate"`
	EndDate   string  `json:"endDate"`
	Address   Address `json:"address"`
	Club      Club    `json:"club"`
	Rules     *Rules  `json:"rules,omitempty"`
	Tables    []Table `json:"tables"`
	Status    int     `json:"status"`
}
