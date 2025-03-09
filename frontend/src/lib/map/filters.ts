import { Tournament } from "../api/types"
import { getCurrentSeasonYears } from "../utils/season"

export const getMapFilters = (tournaments: Tournament[]) => {
    const { seasonStartYear, seasonEndYear } = getCurrentSeasonYears()
    
    return [
        {
            id: 'date_filter',
            dataId: ['current_tournaments'],
            name: [`Tournois à venir pour la saison en cours (${seasonStartYear}-${seasonEndYear})`],
            type: 'timeRange',
            value: [
                Math.min(...tournaments.map(t => new Date(t.startDate).getTime())),
                Math.max(...tournaments.map(t => new Date(t.endDate).getTime()))
            ],
            enlarged: true,
            plotType: 'histogram',
            layerId: ['current_tournaments'],
            field: {
                type: 'timestamp',
                name: `Tournois à venir pour la saison en cours (${seasonStartYear}-${seasonEndYear})`
            }
        },
        {
            id: 'region_filter',
            dataId: ['current_tournaments'],
            name: ['Région'],
            type: 'multiSelect',
            value: [],
            enlarged: false,
            plotType: 'histogram',
            layerId: undefined,
            field: {
                type: 'string',
                name: 'Région',
                defaultValue: [
                    'Auvergne-Rhône-Alpes',
                    'Bourgogne-Franche-Comté',
                    'Bretagne',
                    'Centre-Val de Loire',
                    'Corse',
                    'Grand Est',
                    'Hauts-de-France',
                    'Île-de-France',
                    'Normandie',
                    'Nouvelle-Aquitaine',
                    'Occitanie',
                    'Pays de la Loire',
                    'Provence-Alpes-Côte d\'Azur'
                ]
            }
        },
        {
            id: 'postcode_filter',
            dataId: ['current_tournaments'],
            name: ['Code postal'],
            type: 'input',
            fieldType: 'string',
            value: '',
            enlarged: false,
            layerId: undefined,
            field: {
                type: 'string',
                name: 'Code postal'
            }
        },
        {
            id: 'city_filter',
            dataId: ['current_tournaments'],
            name: ['Ville'],
            type: 'multiSelect',
            value: [],
            enlarged: false,
            field: {
                type: 'string',
                name: 'Ville'
            }
        },
        {
            id: 'club_filter',
            dataId: ['current_tournaments'],
            name: ['Club organisateur'],
            type: 'multiSelect',
            value: [],
            enlarged: false,
            plotType: 'histogram',
            layerId: undefined,
            field: {
                type: 'string',
                name: 'Club organisateur'
            }
        },
        {
            id: 'name_filter',
            dataId: ['current_tournaments'],
            name: ['Nom du tournoi'],
            type: 'input',
            value: '',
            enlarged: false,
            plotType: 'histogram',
            layerId: undefined,
            field: {
                type: 'string',
                name: 'Nom du tournoi'
            }
        },
        {
            id: 'type_filter',
            dataId: ['current_tournaments'],
            name: ['Type de tournoi'],
            type: 'select',
            value: [],
            enlarged: false,
            plotType: 'histogram',
            layerId: undefined,
            field: {
                type: 'string',
                name: 'Type de tournoi'
            }
        },
    ]
}