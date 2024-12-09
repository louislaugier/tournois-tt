package address

// GetFacilityPrefixes returns a list of common facility name prefixes
func GetFacilityPrefixes() []string {
	return []string{
		// Basic terms
		"gymnase", "salle", "complexe", "espace", "stade", "stades", "centre", "palais",

		// Gymnase variations
		"gymnase du", "gymnase de", "gymnase de la", "gymnase des", "gymnase municipal", "gymnase municipal du",
		"gymnase municipal de", "gymnase municipal de la", "gymnase municipal des",
		"gymnase intercommunal", "gymnase intercommunal du", "gymnase intercommunal de",
		"gymnase multisports", "gymnase multi sports", "gymnase multi-sports",

		// Salle variations
		"salle de tennis de table", "salle des sports", "salle du sport", "salle de sport",
		"salle multisports", "salle multi sports", "salle multi-sports",
		"salle multisports du", "salle multisports de", "salle multisports de la", "salle multisports des",
		"salle municipale", "salle municipale du", "salle municipale de", "salle municipale des",
		"salle polyvalente", "salle polyvalente du", "salle polyvalente de", "salle polyvalente des",
		"salle omnisport", "salle omnisports", "salle omnisports du", "salle omnisports de", "salle omnisports des",
		"salle intercommunale", "salle intercommunale du", "salle intercommunale de",
		"salle specialisee", "salle spécialisée", "salle specifique", "salle spécifique",
		"salle de", "salle du", "salle de la", "salle des",

		// Complexe variations
		"complexe sportif", "complexe sportif du", "complexe sportif de", "complexe sportif de la", "complexe sportif des",
		"complexe municipal", "complexe municipal du", "complexe municipal de",
		"complexe multisports", "complexe multi sports", "complexe multi-sports",
		"complexe du", "complexe de", "complexe de la", "complexe des",

		// Espace variations
		"espace sportif", "espace sportif du", "espace sportif de", "espace sportif de la",
		"espace multisports", "espace multi sports", "espace multi-sports",
		"espace municipal", "espace municipal du", "espace municipal de",
		"espace du", "espace de", "espace de la", "espace des",

		// Centre variations
		"centre sportif", "centre sportif du", "centre sportif de", "centre sportif de la",
		"centre municipal", "centre municipal du", "centre municipal de",
		"centre multisports", "centre multi sports", "centre multi-sports",
		"centre du", "centre de", "centre de la", "centre des",

		// Stade variations
		"stade municipal", "stade municipal du", "stade municipal de", "stade municipal de la", "stade municipal des",
		"stade intercommunal", "stade intercommunal du", "stade intercommunal de",
		"stade multisports", "stade multi sports", "stade multi-sports",
		"stade du", "stade de", "stade de la", "stade des",
		"stades municipaux", "stades intercommunaux", "stades multisports",
		"stades du", "stades de", "stades de la", "stades des",

		// Palais variations
		"palais des sports", "palais du sport", "palais de sport",
		"palais municipal", "palais municipal du", "palais municipal de",
		"palais omnisports", "palais omnisport",

		// Special cases
		"maison des sports", "pole sportif", "pôle sportif",
		"cosec", "cosec du", "cosec de", "cosec de la", "cosec des",
		"dojo", "dojo du", "dojo de", "dojo de la", "dojo des",
	}
}
