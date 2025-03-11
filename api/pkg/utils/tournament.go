package utils

// MapTournamentType converts single letter type to full name
func MapTournamentType(t string) string {
	switch t {
	case "I":
		return "International"
	case "A":
		return "National A"
	case "B":
		return "National B"
	case "R":
		return "Régional"
	case "D":
		return "Départemental"
	case "P":
		return "Promotionnel"
	default:
		return t
	}
}
