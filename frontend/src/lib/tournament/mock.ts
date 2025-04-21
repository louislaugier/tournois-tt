import { mockPastTournaments, mockPastCurrentTournaments, mockTournaments } from "../../assets/mockTournaments";
import { Tournament } from "../api/types";
import { normalizeDate } from "../utils/date";
import { getCurrentSeasonStartDate, getLastCompletedSeasonDates } from "../utils/season";

export function upsertMocksIntoDataset(tournamentData: Tournament[], pastCurrentTournamentData: Tournament[], pastTournamentData: Tournament[]) {
    // Handle null/undefined inputs with nullish coalescing
    const currentData = tournamentData ?? [];
    const pastCurrentData = pastCurrentTournamentData ?? [];
    const pastData = pastTournamentData ?? [];
    
    // Get the current season start date and the last completed season date range
    const currentSeasonStartDate = getCurrentSeasonStartDate();
    const { lastCompletedSeasonStartDate, lastCompletedSeasonEndDate } = getLastCompletedSeasonDates();
    
    // Helper function to check if a tournament already exists in the dataset
    const tournamentExists = (dataset: Tournament[], mockTournament: Tournament) => 
        dataset.some(t => 
            t.club.identifier === mockTournament.club.identifier &&
            normalizeDate(t.startDate) === normalizeDate(mockTournament.startDate)
        );
    
    // Helper function to check if a tournament is in the current season
    const isInCurrentSeason = (tournament: Tournament) => {
        const tournamentStartDate = new Date(tournament.startDate);
        return tournamentStartDate >= currentSeasonStartDate;
    };
    
    // Helper function to check if a tournament is in the past completed season
    const isInPastCompletedSeason = (tournament: Tournament) => {
        const tournamentStartDate = new Date(tournament.startDate);
        return tournamentStartDate >= lastCompletedSeasonStartDate && 
               tournamentStartDate <= lastCompletedSeasonEndDate;
    };
    
    // Add new mock tournaments that don't exist in the datasets and match the current season dates
    const tournamentDataWithMocks = [
        ...currentData,
        ...mockTournaments
            .filter(mock => !tournamentExists(currentData, mock))
            .filter(isInCurrentSeason)
    ];

    const pastCurrentTournamentDataWithMocks = [
        ...pastCurrentData,
        ...mockPastCurrentTournaments
            .filter(mock => !tournamentExists(pastCurrentData, mock))
            .filter(isInCurrentSeason)
    ];

    const pastTournamentDataWithMocks = [
        ...pastData,
        ...mockPastTournaments
            .filter(mock => !tournamentExists(pastData, mock))
            .filter(isInPastCompletedSeason)
    ];

    return { tournamentDataWithMocks, pastCurrentTournamentDataWithMocks, pastTournamentDataWithMocks };
}