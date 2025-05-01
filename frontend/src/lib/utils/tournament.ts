/**
 * Maps tournament type codes to their full names
 * Matches the backend implementation in api/pkg/utils/tournament.go
 */
export const mapTournamentType = (type: string): string => {
    switch (type) {
        case "I":
            return "International";
        case "A":
            return "National A";
        case "B":
            return "National B";
        case "R":
            return "Régional";
        case "D":
            return "Départemental";
        case "P":
            return "Promotionnel";
        default:
            return type;
    }
}; 