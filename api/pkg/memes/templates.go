package memes

// MemeTemplate represents a meme text template
type MemeTemplate struct {
	ID       string
	Text     string
	Category string
	Style    string // "panic", "frustration", "celebration", "relatable", etc.
}

// GetAllTemplates returns all meme text templates
func GetAllTemplates() []MemeTemplate {
	return []MemeTemplate{
		// FFTT/Homologation
		{ID: "waiting_homologation", Text: "Quand t'attends l'homologation\nde ton tournoi depuis 3 semaines", Category: "fftt", Style: "waiting"},
		{ID: "refresh_poona", Text: "POV: Tu refresh Poona\ntoutes les 5 minutes\npour voir ton classement", Category: "fftt", Style: "obsessive"},
		{ID: "check_classement", Text: "Moi qui vérifie mon classement\nsur Poona tous les matins", Category: "fftt", Style: "relatable"},
		{ID: "points_calculation", Text: "Moi en train de calculer\ncombien de points il me faut\npour me qualifier", Category: "fftt", Style: "confused"},
		{ID: "gain_12_points", Text: "*Gagne un tournoi*\n\nFFTT: +12 points, félicitations", Category: "fftt", Style: "disappointed"},
		{ID: "lose_87_points", Text: "*Perd un match de poule*\n\nFFTT: -87 points", Category: "fftt", Style: "panic"},
		{ID: "500_francais", Text: "Moi: 500ème français\n\nAussi moi: \"Je suis top 500\"", Category: "fftt", Style: "cope"},
		{ID: "poona_bug", Text: "Poona qui bug\nau pire moment", Category: "fftt", Style: "frustration"},

		// Match situations
		{ID: "gratte_10_9", Text: "Quand ton adversaire\ngratte à 10-9", Category: "match", Style: "panic"},
		{ID: "rate_service_10_10", Text: "Quand tu rates ton service\nà 10-10", Category: "match", Style: "panic"},
		{ID: "faute_arbitre", Text: "POV: L'arbitre appelle \"Faute\"\nsur ton meilleur point", Category: "match", Style: "frustration"},
		{ID: "adversaire_15", Text: "Quand tu réalises que\nton prochain adversaire\nest classé 15 et toi 30", Category: "match", Style: "panic"},
		{ID: "edge_contre_toi", Text: "POV: 5 edges contre toi\ndans le même match", Category: "match", Style: "rage"},
		{ID: "perd_contre_piqueur", Text: "Moi après avoir perdu\ncontre un piqueur", Category: "match", Style: "frustration"},
		{ID: "filet_ton_cote", Text: "Quand la balle touche le filet\net retombe de TON côté", Category: "match", Style: "frustration"},
		{ID: "perd_9_5", Text: "POV: Tu perds 11-9\naprès avoir mené 9-5", Category: "match", Style: "devastated"},
		{ID: "faute_directe", Text: "Quand tu fais une faute directe\nsur une balle facile", Category: "match", Style: "facepalm"},

		// Competitive
		{ID: "bat_le_mec", Text: "Quand tu bats le mec qui t'a\nbattu la dernière fois", Category: "competitive", Style: "celebration"},
		{ID: "elimine_favori", Text: "POV: Tu viens d'éliminer\nle favori du tournoi", Category: "competitive", Style: "celebration"},
		{ID: "crie_cho", Text: "Moi en train de crier \"CHO\"\naprès chaque point", Category: "competitive", Style: "hype"},
		{ID: "remonte_0_2", Text: "POV: Tu remontes de 0-2 en sets", Category: "competitive", Style: "comeback"},
		{ID: "ace_10_10", Text: "Quand tu fais un ace\nà 10-10", Category: "competitive", Style: "celebration"},

		// Club life
		{ID: "president_benevoles", Text: "Le moment où le président\ndu club demande des bénévoles\npour arbitrer", Category: "club", Style: "awkward"},
		{ID: "ramener_ballons", Text: "POV: C'est ton tour de ramener\nles ballons après l'entraînement", Category: "club", Style: "resigned"},
		{ID: "juste_une_derniere", Text: "Quand quelqu'un propose\nde jouer \"juste une dernière\"\net c'est le meilleur joueur du club", Category: "club", Style: "nervous"},
		{ID: "tables_occupees", Text: "Moi qui arrive à l'entraînement\n\nLes 3 tables occupées par\nles mêmes 6 personnes depuis 1h", Category: "club", Style: "waiting"},
		{ID: "club_monte_n3", Text: "Quand le club monte en N3\net tout le monde devient sérieux", Category: "club", Style: "transformation"},
		{ID: "cle_gymnase", Text: "Le mec qui a sa propre clé\ndu gymnase", Category: "club", Style: "legend"},

		// Championship
		{ID: "perd_point_decisif", Text: "Moi: \"C'est qu'un match de poule D2\"\n\n*Perd le point décisif*\n\nTout le club me déteste maintenant", Category: "championship", Style: "regret"},
		{ID: "capitaine_appelle", Text: "Quand t'es N3 et le capitaine\nt'appelle pour jouer N2\n\"juste pour dépanner\"", Category: "championship", Style: "panic"},
		{ID: "jouer_barragiste", Text: "POV: Tu dois jouer le barragiste\net t'es le dernier point", Category: "championship", Style: "pressure"},
		{ID: "coequipier_perd_3_0", Text: "Quand ton coéquipier perd 3-0\ncontre le dernier du classement", Category: "championship", Style: "facepalm"},
		{ID: "stress_capitaine", Text: "Le stress du capitaine qui fait\nla compo 10 minutes avant le match", Category: "championship", Style: "panic"},

		// Money/Registration
		{ID: "economise_argent", Text: "Moi: \"J'économise de l'argent\"\n\n*Voit un tournoi à 1000€*\n\nMoi: \"15€ d'inscription c'est rien\"", Category: "money", Style: "cope"},
		{ID: "paie_pour_perdre", Text: "Quand tu paies 20€\npour perdre au premier tour\nen 10 minutes", Category: "money", Style: "pain"},
		{ID: "budget_raquette", Text: "Mon budget raquette: 300€\nMon budget tournois: 200€\nMon budget gains: 0€\n\nÇa passe", Category: "money", Style: "cope"},

		// Equipment
		{ID: "touche_raquette", Text: "Quand quelqu'un touche\nta raquette sans demander", Category: "equipment", Style: "anger"},
		{ID: "teste_plaques", Text: "Moi qui teste 50 plaques\ndifférentes pour gagner\n1 point de plus", Category: "equipment", Style: "obsessive"},
		{ID: "nouvelles_plaques", Text: "POV: Tu viens de coller\nde nouvelles plaques\net tu perds pire", Category: "equipment", Style: "confused"},
		{ID: "raquette_2003", Text: "Quand ton adversaire a\nune raquette de 2003\net te détruit quand même", Category: "equipment", Style: "embarrassed"},
		{ID: "plaque_decolle", Text: "Quand ta plaque se décolle\nen plein match", Category: "equipment", Style: "panic"},

		// Laziness
		{ID: "entrainer_5_fois", Text: "Moi: \"Cette saison je vais\nm'entraîner 5 fois par semaine\"\n\n*2 semaines plus tard*\n\nMoi sur le canapé", Category: "laziness", Style: "relatable"},
		{ID: "reveil_6h", Text: "Le réveil: 6h pour le tournoi\n\nMoi: \"Pourquoi je m'inscris?\"", Category: "laziness", Style: "regret"},
		{ID: "tutos_youtube_2h", Text: "Moi qui regarde des tutos YouTube\nà 2h du mat au lieu de dormir\navant le tournoi", Category: "laziness", Style: "procrastination"},

		// Déplacements
		{ID: "3h_route_perdre", Text: "Moi qui fais 3h de route\npour perdre en 15 minutes", Category: "travel", Style: "pain"},
		{ID: "covoiturage_apres_perte", Text: "POV: Covoiturage retour\naprès avoir perdu au 1er tour", Category: "travel", Style: "awkward"},
		{ID: "6h_autoroute", Text: "Moi à 6h du mat sur l'autoroute\nen me demandant pourquoi\nje fais ce sport", Category: "travel", Style: "existential"},
		{ID: "temps_route", Text: "Quand tu réalises que tu passes\nplus de temps sur la route\nqu'à jouer", Category: "travel", Style: "realization"},

		// Tournois
		{ID: "scroll_tournois", Text: "Moi qui scroll tournois-tt.fr\nà 2h du mat pour trouver\nun tournoi ce weekend", Category: "tournament", Style: "obsessive"},
		{ID: "tournois_complets", Text: "POV: Tous les tournois du dimanche\nsont complets depuis 2 semaines", Category: "tournament", Style: "frustration"},
		{ID: "inscrit_300km", Text: "Moi qui m'inscris à un tournoi\nà 300km parce que mon régional\nest full", Category: "tournament", Style: "desperate"},
		{ID: "national_tire_1", Text: "Moi qui m'inscris à un National A\nen me disant \"on verra bien\"\n\n*Tire le N°1*\n\nMoi: Pourquoi je fais ça", Category: "tournament", Style: "regret"},

		// Classement
		{ID: "passe_enfin_15", Text: "Quand tu passes enfin 15", Category: "classement", Style: "celebration"},
		{ID: "2_points_monter", Text: "POV: T'étais à 2 points de monter\net tu perds au premier tour", Category: "classement", Style: "devastated"},
		{ID: "top_500", Text: "Moi: 506ème français\nAussi moi: \"Je suis top 500\"", Category: "classement", Style: "cope"},
		{ID: "redescend_classement", Text: "Quand tu redescends de classement\naprès 1 mauvais tournoi", Category: "classement", Style: "pain"},

		// Community
		{ID: "croise_battu_11_0", Text: "Quand tu croises quelqu'un\nque t'as battu 11-0 au supermarché", Category: "community", Style: "awkward"},
		{ID: "pas_vrai_sport", Text: "Quand quelqu'un dit\n\"Le ping c'est pas un vrai sport\"", Category: "community", Style: "anger"},
		{ID: "ping_pong", Text: "Le non-pongiste:\n\"Tu joues au ping pong?\"\n\nMoi: \"TENNIS DE TABLE\"", Category: "community", Style: "triggered"},

		// Saison
		{ID: "saison_progression", Text: "Septembre: \"Cette saison\nje vais monter de classement\"\n\nJuin: -30 points", Category: "season", Style: "disappointment"},
		{ID: "treve_hivernale", Text: "POV: C'est la trêve hivernale\net tu sais plus quoi faire de ta vie", Category: "season", Style: "lost"},
	}
}

// GetTemplatesByCategory returns templates filtered by category
func GetTemplatesByCategory(category string) []MemeTemplate {
	templates := GetAllTemplates()
	var filtered []MemeTemplate
	for _, t := range templates {
		if t.Category == category {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// GetTemplatesByStyle returns templates filtered by style
func GetTemplatesByStyle(style string) []MemeTemplate {
	templates := GetAllTemplates()
	var filtered []MemeTemplate
	for _, t := range templates {
		if t.Style == style {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// GetRandomTemplate returns a random template
func GetRandomTemplate() MemeTemplate {
	templates := GetAllTemplates()
	// TODO: Add proper random selection
	return templates[0]
}

