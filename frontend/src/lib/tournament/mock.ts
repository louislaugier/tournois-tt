import { mockTournaments } from "../../assets/mockTournaments";
import { Tournament } from "../api/types";
import { normalizeDate } from "../utils/date";

export function upsertMocksIntoDataset(tournamentData: Tournament[]): Tournament[] {
    let tournamentDataWithMocks = tournamentData || [];

    mockTournaments.forEach(mockTournament => {
        // Add mock tournament to tournament data only if it's not already in the data (check club identifier & if on the same date)
        const existingTournament = tournamentDataWithMocks.find(t =>
            t.club.identifier === mockTournament.club.identifier &&
            normalizeDate(t.startDate) === normalizeDate(mockTournament.startDate)
        );
        if (!existingTournament) {
            tournamentDataWithMocks = [...tournamentDataWithMocks, mockTournament];
        }
    });

    return tournamentDataWithMocks
}