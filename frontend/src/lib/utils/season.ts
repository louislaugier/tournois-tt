import { getCurrentMonth, getCurrentYear } from "./date";

export function getLastCompletedSeasonDates() {
    const { seasonStartYear } = getCurrentSeasonYears()

    // For past tournaments, we want the season BEFORE the current season
    const pastSeasonStartYear = seasonStartYear - 1;
    const pastSeasonEndYear = seasonStartYear;

    const lastCompletedSeasonStartDate = new Date(pastSeasonStartYear, 6, 1); // July 1st of past season start year
    lastCompletedSeasonStartDate.setHours(0, 0, 0, 0);

    const lastCompletedSeasonEndDate = new Date(pastSeasonEndYear, 5, 30); // June 30th of past season end year
    lastCompletedSeasonEndDate.setHours(23, 59, 59, 999);

    return { lastCompletedSeasonStartDate, lastCompletedSeasonEndDate }
}

export function getCurrentSeasonStartDate() {
    const { seasonStartYear } = getCurrentSeasonYears()

    return new Date(seasonStartYear, 6, 1); // July 1st of current season start year
}

export function getCurrentSeasonYears() {
    // Get the current date and determine the latest finished season
    const currentMonth = getCurrentMonth()
    const currentYear = getCurrentYear()
    let seasonStartYear, seasonEndYear;
    if (currentMonth >= 7) { // July or later
        seasonStartYear = currentYear;
        seasonEndYear = currentYear + 1;
    } else {
        seasonStartYear = currentYear - 1;
        seasonEndYear = currentYear;
    }

    return { seasonStartYear, seasonEndYear }
}
