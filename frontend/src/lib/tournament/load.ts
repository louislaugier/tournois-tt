import { TournamentQueryBuilder } from "../api/tournaments";
import { Tournament } from "../api/types";
import { getTodayMidnight, getYesterdayMidnight } from "../utils/date";
import { getCurrentSeasonStartDate, getLastCompletedSeasonDates } from "../utils/season";
import { upsertMocksIntoDataset } from "./mock";

export const loadTournaments = async (setIsLoading: (value: React.SetStateAction<boolean>) => void, setCurrentTournaments: (value: React.SetStateAction<Tournament[]>) => void, setPastCurrentTournaments: (value: React.SetStateAction<Tournament[]>) => void, setPastTournaments: (value: React.SetStateAction<Tournament[]>) => void) => {
    try {
        const query = new TournamentQueryBuilder()
            .startDateRange(getYesterdayMidnight())
            .orderByStartDate('asc')
            .itemsPerPage(999999);

        const currentTournamentsData = await query.executeAndLogAll();

        setCurrentTournaments(upsertMocksIntoDataset(currentTournamentsData));

        const pastCurrentTournamentsQuery = new TournamentQueryBuilder()
            .startDateRange(getCurrentSeasonStartDate(), getTodayMidnight())
            .orderByStartDate('asc')
            .itemsPerPage(999999);

        const pastCurrentTournamentsData = await pastCurrentTournamentsQuery.executeAndLogAll();

        setPastCurrentTournaments(pastCurrentTournamentsData || []);

        const { lastCompletedSeasonStartDate, lastCompletedSeasonEndDate } = getLastCompletedSeasonDates()

        const pastTournamentsQuery = new TournamentQueryBuilder()
            .startDateRange(lastCompletedSeasonStartDate, lastCompletedSeasonEndDate)
            .orderByStartDate('asc')
            .itemsPerPage(999999);

        const pastTournamentsData = await pastTournamentsQuery.executeAndLogAll();

        setPastTournaments(pastTournamentsData || []);
    } catch (error) {
        console.error('Failed to load tournaments:', error);
    } finally {
        setIsLoading(false);
    }
};