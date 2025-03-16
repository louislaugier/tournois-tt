import { TournamentQueryBuilder } from "../api/tournaments";
import { Tournament } from "../api/types";
import { getTodayMidnight, getYesterdayMidnight } from "../utils/date";
import { getCurrentSeasonStartDate, getLastCompletedSeasonDates } from "../utils/season";
import { upsertMocksIntoDataset } from "./mock";

export const loadTournaments = async (setIsLoading: (value: React.SetStateAction<boolean>) => void, setCurrentTournaments: (value: React.SetStateAction<Tournament[]>) => void, setPastCurrentTournaments: (value: React.SetStateAction<Tournament[]>) => void, setPastTournaments: (value: React.SetStateAction<Tournament[]>) => void) => {
    try {
        const { lastCompletedSeasonStartDate, lastCompletedSeasonEndDate } = getLastCompletedSeasonDates();

        const query = new TournamentQueryBuilder()
            .startDateRange(lastCompletedSeasonStartDate, undefined) // Fetch ALL tournaments at once
            .orderByStartDate('asc')
            .itemsPerPage(999999);

        const allTournamentsData = await query.executeAndLogAll();

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
        const {tournamentDataWithMocks, pastTournamentDataWithMocks} = upsertMocksIntoDataset(currentTournamentsData, pastTournamentsData);
        
        setCurrentTournaments(tournamentDataWithMocks);

        setPastCurrentTournaments(pastCurrentTournamentsData);

        setPastTournaments(pastTournamentDataWithMocks);
    } catch (error) {
        console.error('Failed to load tournaments:', error);
    } finally {
        setIsLoading(false);
    }
};