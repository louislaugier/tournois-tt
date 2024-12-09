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
	Status    int     `json:"status"`
}
