import { Tournament } from "../api/types";
import { getMapFilters } from "./filters";
import { getMapLayers } from "./layers";

export const tooltipFields = [
    { name: 'Nom du tournoi', format: null },
    { name: 'Type de tournoi', format: null },
    { name: 'Club organisateur', format: null },
    { name: 'Dotation totale (€)', format: null },
    { name: 'Date(s)', format: null },
    { name: 'Adresse', format: null },
    { name: 'Règlement', format: null },
    { name: 'Inscription', format: null },
]

export const getMapConfig = (tournaments: Tournament[]) => {
    return {
        visState: {
            interactionConfig: {
                tooltip: {
                    fieldsToShow: {
                        current_tournaments: tooltipFields,
                        past_current_tournaments: tooltipFields,
                        past_tournaments: tooltipFields,
                        france: []
                    },
                    enabled: true,
                    compareMode: false,
                    compareType: 'absolute'
                },
                brush: {
                    enabled: false
                },
                coordinate: {
                    enabled: false
                },
                geocoder: {
                    enabled: false
                }
            },
            layers: getMapLayers(),
            filters: getMapFilters(tournaments),
        },
        mapState: {
            latitude: 46.777138,
            longitude: 2.804568,
            pitch: 0,
            bearing: 0,
            zoom: 5.6,
            dragRotate: false
        },
        mapStyle: {
            visibleLayerGroups: {
                label: true,
                road: true,
                building: true,
                water: true,
                land: true,
                border: true
            }
        }
    }
}