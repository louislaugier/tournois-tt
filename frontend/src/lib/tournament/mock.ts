import { mockPastTournaments, mockTournaments } from "../../assets/mockTournaments";
import { Tournament } from "../api/types";
import { normalizeDate } from "../utils/date";

export function upsertMocksIntoDataset(tournamentData: Tournament[], pastTournamentData: Tournament[]) {
    // Handle null/undefined inputs with nullish coalescing
    const currentData = tournamentData ?? [];
    const pastData = pastTournamentData ?? [];
    
    // Helper function to check if a tournament already exists in the dataset
    const tournamentExists = (dataset: Tournament[], mockTournament: Tournament) => 
        dataset.some(t => 
            t.club.identifier === mockTournament.club.identifier &&
            normalizeDate(t.startDate) === normalizeDate(mockTournament.startDate)
        );
    
    // Add new mock tournaments that don't exist in the datasets
    const tournamentDataWithMocks = [
        ...currentData,
        ...mockTournaments.filter(mock => !tournamentExists(currentData, mock))
    ];
    
    const pastTournamentDataWithMocks = [
        ...pastData,
        ...mockPastTournaments.filter(mock => !tournamentExists(pastData, mock))
    ];

    return { tournamentDataWithMocks, pastTournamentDataWithMocks };
}