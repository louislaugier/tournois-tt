import { formatPostcode, formatCityName, getRegionFromPostalCode } from '../utils/address';
import { getTodayMidnight, formatDateDDMMYYYY } from '../utils/date';
import { getCurrentSeasonYears } from '../utils/season';
import { mapTournamentType } from '../utils/tournament';
// import franceBordersRaw from './../../assets/metropole-et-outre-mer.json';

const { seasonStartYear, seasonEndYear } = getCurrentSeasonYears()

export const tournamentFields = [
    { name: 'latitude', type: 'real' },
    { name: 'longitude', type: 'real' },
    { name: 'Nom du tournoi', type: 'string' },
    { name: 'Type de tournoi', type: 'string' },
    { name: 'Club organisateur', type: 'string' },
    { name: 'Dotation totale (€)', type: 'real', analyzerType: 'INT' },
    { name: 'Date(s)', type: 'date' },
    { name: `Tournois à venir pour la saison en cours (${seasonStartYear}-${seasonEndYear})`, type: 'date' },
    { name: 'Adresse', type: 'string' },
    { name: 'Règlement', type: 'string' },
    { name: 'Inscription', type: 'string' },
    { name: 'Code postal', type: 'string' },
    { name: 'Ville', type: 'string' },
    { name: 'Région', type: 'string' },
    { name: 'count', type: 'string' },
]

export const getTournamentRows = (allTournamentsForMap: any) => {
    return Object.values(
        allTournamentsForMap.reduce((acc, t) => {
            const key = `${t.address.latitude},${t.address.longitude}`;
            if (!acc[key]) acc[key] = {
                latitude: t.address.latitude,
                longitude: t.address.longitude,
                tournaments: [],
                count: 0
            };

            acc[key].tournaments.push(t);
            acc[key].count++;

            return acc;
        }, {} as Record<string, any>)
    ).map((location: any) => {
        const postalCode = Array.from(new Set<string>(location.tournaments.map(t => formatPostcode(t.address.postalCode)))).join(' | ');
        return [
            location.latitude,
            location.longitude,
            location.tournaments.map(t => t.name).join(' | '),
            Array.from(new Set<string>(location.tournaments.map(t => mapTournamentType(t.type)))).join(' | '),
            Array.from(new Set<string>(location.tournaments.map(t => `${t.club.name} (${t.club.identifier})`))).join(' | '),
            location.tournaments.map(t => {
                // Ensure we always have a valid numeric value for endowment filtering
                let endowmentValue = 0;
                if (typeof t.endowment === 'number' && t.endowment > 0) {
                    endowmentValue = Math.floor(t.endowment / 100);
                } else if (t.tables && t.tables.length > 0) {
                    // Calculate from tables
                    endowmentValue = Math.floor((t.tables.reduce((sum, table) => sum + (table.endowment || 0), 0) || 0) / 100);
                }
                return endowmentValue;
            }).join(' | '),
            location.tournaments.map(t =>
                `${formatDateDDMMYYYY(t.startDate)}${t.startDate !== t.endDate ? ` au ${formatDateDDMMYYYY(t.endDate)}` : ''}`
            ).join(' | '),
            Math.min(...location.tournaments.map(t => new Date(t.startDate).getTime())), // for date filter
            location.tournaments[0].address.streetAddress
                ? `${location.tournaments[0].address.disambiguatingDescription ? location.tournaments[0].address.disambiguatingDescription + ' ' : ''}${location.tournaments[0].address.streetAddress}, ${location.tournaments[0].address.postalCode} ${location.tournaments[0].address.addressLocality}`
                : 'Adresse non disponible',
            location.tournaments.map(t => {
                // For debugging purposes - console log tournament data to see affiche
                if (t.identifier === 'MOCK-ATT-XV-2025') {
                    console.log('Found mock tournament:', t);
                }
                
                // Display rules URL if available
                if (t.rules?.url) {
                    return location.count > 1 ? t.rules.url.replace('https://', '') : t.rules.url;
                }
                
                // If affiche is available, display it
                if (t.affiche) {
                    const afficheDisplay = location.count > 1 ? t.affiche.replace('https://', '') : t.affiche;
                    return t.rules ? afficheDisplay : `Pas de règlement (Affiche: ${afficheDisplay})`;
                }
                
                // Default case: no rules, no affiche
                return 'Pas de règlement';
            }).join(' | '),
            location.tournaments.map(t => {
                // Display signup URL if available
                if (t.signupUrl) {
                    return location.count > 1 ? t.signupUrl.replace('https://', '') : t.signupUrl;
                }
                return '/';
            }).join(' | '),
            postalCode,
            formatCityName(location.tournaments[0].address.addressLocality),
            getRegionFromPostalCode(postalCode.split(' | ')[0]),
            location.count > 1 ? location.count.toString() : '',
        ]
    })
}

// export const franceBorder = {
//     info: {
//         label: 'France',
//         id: 'france'
//     },
//     data: {
//         fields: [
//             { name: '_geojson', type: 'geojson' }
//         ],
//         rows: [[typeof franceBordersRaw === 'string' ? JSON.parse(franceBordersRaw) : franceBordersRaw]]
//     }
// }