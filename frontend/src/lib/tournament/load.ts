import { TournamentQueryBuilder } from "../api/tournaments";
import { Tournament } from "../api/types";
import { getTodayMidnight, getYesterdayMidnight, normalizeDate } from "../utils/date";
import { getCurrentSeasonStartDate, getLastCompletedSeasonDates } from "../utils/season";
import { upsertMocksIntoDataset } from "./mock";

// Deduplicate tournaments with the same club and start date
function deduplicateTournaments(tournaments: Tournament[]): Tournament[] {
    // Create a map to track unique tournaments by club+date
    const uniqueTournaments = new Map<string, Tournament>();
    
    // Loop through all tournaments
    for (const tournament of tournaments) {
        // Create a unique key based on club identifier and start date
        const key = `${tournament.club.identifier}|${normalizeDate(tournament.startDate)}`;
        
        // If we've never seen this club+date before, or this tournament has a higher ID (newer),
        // add/update it in our map
        if (!uniqueTournaments.has(key) || tournament.id > uniqueTournaments.get(key)!.id) {
            uniqueTournaments.set(key, tournament);
        }
    }
    
    // Convert map back to array
    return Array.from(uniqueTournaments.values());
}

export const loadTournaments = async (setIsLoading: (value: React.SetStateAction<boolean>) => void, setCurrentTournaments: (value: React.SetStateAction<Tournament[]>) => void, setPastCurrentTournaments: (value: React.SetStateAction<Tournament[]>) => void, setPastTournaments: (value: React.SetStateAction<Tournament[]>) => void) => {
    try {
        const { lastCompletedSeasonStartDate, lastCompletedSeasonEndDate } = getLastCompletedSeasonDates();

        const query = new TournamentQueryBuilder()
            .startDateRange(lastCompletedSeasonStartDate, undefined) // Fetch ALL tournaments at once
            .orderByStartDate('asc')
            .itemsPerPage(999999);

        const allTournamentsData = await query.executeAndLogAll();
        
        // Debug: Check if Draveil tournament has page field from API
        const draveilFromAPI = allTournamentsData.filter(t => t.name && t.name.includes('DRAVEIL'));
        if (draveilFromAPI.length > 0) {
            console.log('🔍 DRAVEIL from API response:', draveilFromAPI.map(t => ({
                id: t.id,
                name: t.name,
                page: t.page,
                hasPageField: 'page' in t,
                pageValue: t.page
            })));
        }

        // Get date ranges
        const yesterdayMidnight = getYesterdayMidnight();
        const todayMidnight = getTodayMidnight();
        const currentSeasonStartDate = getCurrentSeasonStartDate();

        // Filter current tournaments (from yesterday onwards)
        const currentTournamentsData = allTournamentsData.filter(tournament => {
            const tournamentStartDate = new Date(tournament.startDate);
            return tournamentStartDate >= yesterdayMidnight;
        });

        // Filter past current tournaments (within current season)
        const pastCurrentTournamentsData = allTournamentsData.filter(tournament => {
            const tournamentStartDate = new Date(tournament.startDate);
            return tournamentStartDate >= currentSeasonStartDate && 
                   tournamentStartDate <= todayMidnight;
        });

        // Filter past tournaments (within last completed season)
        const pastTournamentsData = allTournamentsData.filter(tournament => {
            const tournamentStartDate = new Date(tournament.startDate);
            return tournamentStartDate >= lastCompletedSeasonStartDate && 
                   tournamentStartDate <= lastCompletedSeasonEndDate;
        });

        // Apply mocks to current tournaments
        const {tournamentDataWithMocks, pastCurrentTournamentDataWithMocks, pastTournamentDataWithMocks} = upsertMocksIntoDataset(currentTournamentsData, pastCurrentTournamentsData, pastTournamentsData);
        
        // Deduplicate at the very end
        const dedupedCurrentTournaments = deduplicateTournaments(tournamentDataWithMocks);
        const dedupedPastCurrentTournaments = deduplicateTournaments(pastCurrentTournamentDataWithMocks);
        const dedupedPastTournaments = deduplicateTournaments(pastTournamentDataWithMocks);
        
        setCurrentTournaments(dedupedCurrentTournaments);
        setPastCurrentTournaments(dedupedPastCurrentTournaments);
        setPastTournaments(dedupedPastTournaments);
    } catch (error) {
        console.error('Failed to load tournaments:', error);
    } finally {
        setIsLoading(false);
    }
};