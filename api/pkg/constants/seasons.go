package constants

// Season names in French
const (
	// Seasons
	Spring = "printemps"
	Summer = "été"
	Fall   = "automne"
	Winter = "hiver"

	// Holiday periods
	Easter    = "pâques"
	Christmas = "noël"
	NewYear   = "nouvel an"
	AllSaints = "toussaint"
)

// Month names in French
const (
	January   = "janvier"
	February  = "février"
	March     = "mars"
	April     = "avril"
	May       = "mai"
	June      = "juin"
	July      = "juillet"
	August    = "août"
	September = "septembre"
	October   = "octobre"
	November  = "novembre"
	December  = "décembre"
)

// Common tournament name patterns
var (
	// Tournament name patterns by season
	TournamentPatterns = []string{
		"tournoi de " + Spring,
		"tournoi de " + Summer,
		"tournoi de " + Fall,
		"tournoi de " + Winter,
		"tournoi de " + Easter,
		"tournoi de " + Christmas,
		"tournoi de " + NewYear,
		"tournoi de " + AllSaints,
		"tournoi du " + Spring,
		"tournoi du " + Summer,
		"tournoi du " + Fall,
		"tournoi du " + Winter,
		"tournoi d'" + Easter,
		"tournoi de " + Christmas,
		"tournoi du " + NewYear,
		"tournoi de la " + AllSaints,
	}

	// Tournament name patterns by month
	TournamentMonthPatterns = []string{
		"tournoi de " + January,
		"tournoi de " + February,
		"tournoi de " + March,
		"tournoi de " + April,
		"tournoi de " + May,
		"tournoi de " + June,
		"tournoi de " + July,
		"tournoi de " + August,
		"tournoi de " + September,
		"tournoi de " + October,
		"tournoi de " + November,
		"tournoi de " + December,
		"tournoi du mois de " + January,
		"tournoi du mois de " + February,
		"tournoi du mois de " + March,
		"tournoi du mois de " + April,
		"tournoi du mois de " + May,
		"tournoi du mois de " + June,
		"tournoi du mois de " + July,
		"tournoi du mois de " + August,
		"tournoi du mois de " + September,
		"tournoi du mois de " + October,
		"tournoi du mois de " + November,
		"tournoi du mois de " + December,
	}

	// Tournament name pattern capitals variations
	TournamentPatternsCapitalized = []string{
		"Tournoi de " + Spring,
		"Tournoi de " + Summer,
		"Tournoi de " + Fall,
		"Tournoi de " + Winter,
		"Tournoi de " + Easter,
		"Tournoi de " + Christmas,
		"Tournoi de " + NewYear,
		"Tournoi de " + AllSaints,
		"Tournoi du " + Spring,
		"Tournoi du " + Summer,
		"Tournoi du " + Fall,
		"Tournoi du " + Winter,
		"Tournoi d'" + Easter,
		"Tournoi de " + Christmas,
		"Tournoi du " + NewYear,
		"Tournoi de la " + AllSaints,
	}

	// Tournament month pattern capitals variations
	TournamentMonthPatternsCapitalized = []string{
		"Tournoi de " + January,
		"Tournoi de " + February,
		"Tournoi de " + March,
		"Tournoi de " + April,
		"Tournoi de " + May,
		"Tournoi de " + June,
		"Tournoi de " + July,
		"Tournoi de " + August,
		"Tournoi de " + September,
		"Tournoi de " + October,
		"Tournoi de " + November,
		"Tournoi de " + December,
		"Tournoi du mois de " + January,
		"Tournoi du mois de " + February,
		"Tournoi du mois de " + March,
		"Tournoi du mois de " + April,
		"Tournoi du mois de " + May,
		"Tournoi du mois de " + June,
		"Tournoi du mois de " + July,
		"Tournoi du mois de " + August,
		"Tournoi du mois de " + September,
		"Tournoi du mois de " + October,
		"Tournoi du mois de " + November,
		"Tournoi du mois de " + December,
	}

	// All tournament patterns combined
	AllTournamentPatterns = append(
		append(
			append(TournamentPatterns, TournamentMonthPatterns...),
			TournamentPatternsCapitalized...,
		),
		TournamentMonthPatternsCapitalized...,
	)

	// Special tournament names
	SpecialTournaments = []string{
		"tournoi national",
		"tournoi régional",
		"tournoi international",
		"tournoi open",
		"open",
		"grand prix",
		"critérium",
		"championnat",
	}

	// Special tournament names capitalized
	SpecialTournamentsCapitalized = []string{
		"Tournoi National",
		"Tournoi Régional",
		"Tournoi International",
		"Tournoi Open",
		"Open",
		"Grand Prix",
		"Critérium",
		"Championnat",
	}
)
