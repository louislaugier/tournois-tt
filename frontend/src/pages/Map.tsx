import { KeplerGl } from "@kepler.gl/components";
import React, { useState, useEffect } from "react";
import { Tournament } from "../lib/api/types";
import { initializeDateFormatter } from "../lib/date/dateFormatter";
import { initializeDOMCleaner } from "../lib/domCleaner";
import { franceBorders } from "../lib/map/border";
import { DEFAULT_MAP_CONFIG } from "../lib/map/config";
import { MAPBOX_TOKEN } from "../lib/map/constants";
import { initializeSidebarCustomizer } from "../lib/sidebarCustomizer";
import { store } from "../lib/store";
import { formatPostcode, formatCityName, getRegionFromPostalCode } from "../lib/utils/address";
import { formatDateDDMMYYYY } from "../lib/utils/date";
import fr from "../locales/fr";
import { addDataToMap } from '@kepler.gl/actions';
import { loadTournaments } from "../lib/tournament/load";
import { getLastCompletedSeasonYears } from "../lib/utils/season";

export const Map: React.FC = () => {
    const [tournaments, setTournaments] = useState<Tournament[]>([]);
    const [pastTournaments, setPastTournaments] = useState<Tournament[]>([]);
    const [isLoading, setIsLoading] = useState<boolean>(true);

    useEffect(() => {
        initializeSidebarCustomizer();
        initializeDateFormatter();
        initializeDOMCleaner();
    }, []);

    useEffect(() => {
        loadTournaments(setIsLoading, setTournaments, setPastTournaments);
    }, []);

    useEffect(() => {
        const observer = new MutationObserver((mutations) => {
            mutations.forEach((mutation) => {
                if (mutation.type === 'childList') {
                    // Handle map-popover__layer-name
                    document.querySelectorAll('.map-popover__layer-name').forEach((el) => {
                        if (el.textContent?.includes('Tournoi')) {
                            (el as HTMLElement).style.display = 'none';
                            // Find and style jlDYGb elements within this tooltip
                            const tooltipContainer = el.closest('.map-popover');
                            tooltipContainer?.querySelectorAll('.jlDYGb').forEach((jlElement) => {
                                (jlElement as HTMLElement).style.marginBottom = '12.5px';
                            });
                        } else {
                            (el as HTMLElement).style.display = '';
                            // Reset margin for non-Tournoi tooltips
                            const tooltipContainer = el.closest('.map-popover');
                            tooltipContainer?.querySelectorAll('.jlDYGb').forEach((jlElement) => {
                                (jlElement as HTMLElement).style.marginBottom = '';
                            });
                        }
                    });

                    // Handle kWuINq separately
                    document.querySelectorAll('.kWuINq').forEach((el) => {
                        const tooltipContainer = el.closest('.map-popover');
                        const tournamentName = tooltipContainer?.querySelector('.row__name')?.textContent;
                        if (tournamentName === 'Nom du tournoi') {
                            (el as HTMLElement).style.display = 'none';
                        } else {
                            (el as HTMLElement).style.display = '';
                        }
                    });
                }
            });
        });

        observer.observe(document.body, {
            childList: true,
            subtree: true
        });

        return () => observer.disconnect();
    }, []);

    const { seasonStartYear, seasonEndYear } = getLastCompletedSeasonYears()

    useEffect(() => {
        if (tournaments?.length > 0) {
            const tournamentsWithCoordinates = tournaments.filter(
                t => t.address?.latitude && t.address?.longitude
            );
            const tournamentsWithoutCoordinates = tournaments.filter(
                t => !t.address?.latitude || !t.address?.longitude
            );

            const allTournamentsForMap = [
                ...tournamentsWithCoordinates,
                ...tournamentsWithoutCoordinates.map(t => ({
                    ...t,
                    address: {
                        ...t.address,
                        latitude: 46.777138,
                        longitude: 2.804568,
                        approximate: true
                    }
                }))
            ].map(t => ({
                ...t,
                address: {
                    ...t.address,
                    postalCode: formatPostcode(t.address.postalCode)
                }
            }));

            let allPastTournamentsForMap: Tournament[] = [];
            if (pastTournaments?.length > 0) {
                const pastTournamentsWithCoordinates = pastTournaments.filter(
                    t => t.address?.latitude && t.address?.longitude
                );
                const pastTournamentsWithoutCoordinates = pastTournaments.filter(
                    t => !t.address?.latitude || !t.address?.longitude
                );

                allPastTournamentsForMap = [
                    ...pastTournamentsWithCoordinates,
                    ...pastTournamentsWithoutCoordinates.map(t => ({
                        ...t,
                        address: {
                            ...t.address,
                            latitude: 46.777138,
                            longitude: 2.804568,
                            approximate: true
                        }
                    }))
                ].map(t => ({
                    ...t,
                    address: {
                        ...t.address,
                        postalCode: formatPostcode(t.address.postalCode)
                    }
                }));
            }

            if (allTournamentsForMap.length > 0) {
                const tooltipFields = [
                    { name: 'Nom du tournoi', format: null },
                    { name: 'Type de tournoi', format: null },
                    { name: 'Club organisateur', format: null },
                    { name: 'Dotation totale (€)', format: null },
                    { name: 'Date(s)', format: null },
                    { name: 'Adresse', format: null },
                    { name: 'Règlement', format: null }
                ];

                const enhancedMapConfig = {
                    ...DEFAULT_MAP_CONFIG,
                    visState: {
                        ...DEFAULT_MAP_CONFIG.visState,
                        interactionConfig: {
                            ...DEFAULT_MAP_CONFIG.visState.interactionConfig,
                            tooltip: {
                                fieldsToShow: {
                                    tournoi: tooltipFields,
                                    tournoi_passe: tooltipFields,
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
                        filters: [
                            {
                                id: 'date_filter',
                                dataId: ['tournoi'],
                                name: [`Tournois à venir pour la saison en cours (${seasonStartYear}-${seasonEndYear})`],
                                type: 'timeRange',
                                value: [
                                    Math.min(...tournaments.map(t => new Date(t.startDate).getTime())),
                                    Math.max(...tournaments.map(t => new Date(t.endDate).getTime()))
                                ],
                                enlarged: true,
                                plotType: 'histogram',
                                layerId: undefined,
                                field: {
                                    type: 'timestamp',
                                    name: `Tournois à venir pour la saison en cours (${seasonStartYear}-${seasonEndYear})`
                                }
                            },
                            {
                                id: 'region_filter',
                                dataId: ['tournoi'],
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
                                dataId: ['tournoi'],
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
                                dataId: ['tournoi'],
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
                                dataId: ['tournoi'],
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
                                dataId: ['tournoi'],
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
                                dataId: ['tournoi'],
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
                        ],
                        layers: [
                            {
                                id: 'tournament_points',
                                type: 'point',
                                config: {
                                    dataId: 'tournoi',
                                    label: (() => {
                                        return `Saison en cours (${seasonStartYear}-${seasonEndYear})`;
                                    })(),
                                    color: [31, 186, 214] as [number, number, number],
                                    columns: {
                                        lat: 'latitude',
                                        lng: 'longitude'
                                    } as { [key: string]: string },
                                    isVisible: true,
                                    visConfig: {
                                        radius: 20,
                                        fixedRadius: false,
                                        opacity: 0.8,
                                        outline: false,
                                        filled: true,
                                        radiusRange: [20, 30]
                                    },
                                    textLabel: {
                                        field: { name: 'count_display', type: 'string' },
                                        color: [255, 255, 255] as [number, number, number],
                                        size: 16,
                                        offset: [0, 0] as [number, number],
                                        anchor: 'middle',
                                        alignment: 'center',
                                        background: true,
                                        backgroundColor: [0, 0, 0, 0.5] as [number, number, number, number],
                                        outlineWidth: 0,
                                        outlineColor: [0, 0, 0, 0.5] as [number, number, number, number]
                                    }
                                },
                                visualChannels: {
                                    colorField: null,
                                    colorScale: 'quantile',
                                    sizeField: { name: 'size_multiplier', type: 'real' },
                                    sizeScale: 'linear',
                                    strokeColorField: null,
                                    strokeColorScale: 'quantile'
                                }
                            },
                            {
                                id: 'past_tournament_points',
                                type: 'point',
                                config: {
                                    dataId: 'tournoi_passe',
                                    label: (() => {
                                        const now = new Date();
                                        const currentYear = now.getFullYear();
                                        const currentMonth = now.getMonth() + 1; // JavaScript months are 0-based

                                        let seasonStartYear, seasonEndYear;
                                        if (currentMonth >= 7) { // July or later
                                            seasonStartYear = currentYear;
                                            seasonEndYear = currentYear + 1;
                                        } else {
                                            seasonStartYear = currentYear - 1;
                                            seasonEndYear = currentYear;
                                        }

                                        // For past tournaments, we want the season BEFORE the current season
                                        const pastSeasonStartYear = seasonStartYear - 1;
                                        const pastSeasonEndYear = seasonStartYear;

                                        return `Saison précédente (${pastSeasonStartYear}-${pastSeasonEndYear})`;
                                    })(),
                                    color: [155, 89, 182] as [number, number, number],
                                    columns: {
                                        lat: 'latitude',
                                        lng: 'longitude'
                                    } as { [key: string]: string },
                                    isVisible: true,
                                    visConfig: {
                                        radius: 22,
                                        fixedRadius: false,
                                        opacity: 0.8,
                                        outline: false,
                                        filled: true,
                                        radiusRange: [20, 30]
                                    },
                                    textLabel: {
                                        field: { name: 'count_display', type: 'string' },
                                        color: [255, 255, 255] as [number, number, number],
                                        size: 14,
                                        offset: [0, 0] as [number, number],
                                        anchor: 'middle',
                                        alignment: 'center',
                                        background: true,
                                        backgroundColor: [0, 0, 0, 0.5] as [number, number, number, number],
                                        outlineWidth: 0,
                                        outlineColor: [0, 0, 0, 0.5] as [number, number, number, number]
                                    }
                                },
                                visualChannels: {
                                    colorField: null,
                                    colorScale: 'quantile',
                                    sizeField: { name: 'size_multiplier', type: 'real' },
                                    sizeScale: 'linear',
                                    strokeColorField: null,
                                    strokeColorScale: 'quantile'
                                }
                            },

                        ]
                    },
                    mapState: {
                        ...DEFAULT_MAP_CONFIG.mapState,
                        pitch: 0,
                        bearing: 0,
                        dragRotate: false
                    },
                    mapStyle: {
                        ...DEFAULT_MAP_CONFIG.mapStyle,
                        topLayerGroups: {},
                        visibleLayerGroups: {
                            ...DEFAULT_MAP_CONFIG.mapStyle.visibleLayerGroups,
                            label: true,
                            road: true,
                            building: true,
                            water: true,
                            land: true,
                            border: true
                        }
                    }
                };

                const tournamentFields = [
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
                    { name: 'Code postal', type: 'string' },
                    { name: 'Ville', type: 'string' },
                    { name: 'Région', type: 'string' },
                    { name: 'count', type: 'integer' },
                    { name: 'count_display', type: 'string' },
                    { name: 'size_multiplier', type: 'real' }
                ]
                console.log('Total tournaments before filtering:', allTournamentsForMap.length);
                const tournamentsDataset = {
                    fields: tournamentFields,
                    rows: Object.values(
                        allTournamentsForMap.reduce((acc, t) => {
                            // Skip tournaments that have already ended
                            const endDate = new Date(t.endDate);
                            const now = new Date();
                            // Set both dates to start of day for fair comparison
                            endDate.setHours(0, 0, 0, 0);
                            now.setHours(0, 0, 0, 0);
                            console.log(`Tournament: ${t.name}, End Date: ${endDate.toISOString()}, Now: ${now.toISOString()}, Filtered: ${endDate < now}`);
                            if (endDate < now) {
                                return acc;
                            }

                            const key = `${t.address.latitude},${t.address.longitude}`;
                            if (!acc[key]) {
                                acc[key] = {
                                    latitude: t.address.latitude,
                                    longitude: t.address.longitude,
                                    tournaments: [],
                                    count: 0
                                };
                            }
                            acc[key].tournaments.push(t);
                            acc[key].count++;
                            return acc;
                        }, {} as Record<string, any>)
                    ).map(location => {
                        console.log('Location data:', location);
                        const postalCode = Array.from(new Set<string>(location.tournaments.map(t => formatPostcode(t.address.postalCode)))).join(' | ');
                        return [
                            location.latitude,
                            location.longitude,
                            location.tournaments.map(t => t.name).join(' | '),
                            Array.from(new Set<string>(location.tournaments.map(t => t.type))).join(' | '),
                            Array.from(new Set<string>(location.tournaments.map(t => `${t.club.name} (${t.club.identifier})`))).join(' | '),
                            location.tournaments.map(t =>
                            (typeof t.endowment === 'number' && t.endowment > 0
                                ? Math.floor(t.endowment / 100)
                                : (t.tables?.reduce((sum, table) => sum + (table.endowment || 0), 0) || 0) / 100)
                            ).join(' | '),
                            location.tournaments.map(t =>
                                `${formatDateDDMMYYYY(t.startDate)}${t.startDate !== t.endDate ? ` au ${formatDateDDMMYYYY(t.endDate)}` : ''}`
                            ).join(' | '),
                            Math.min(...location.tournaments.map(t => new Date(t.startDate).getTime())),
                            location.tournaments[0].address.streetAddress
                                ? `${location.tournaments[0].address.disambiguatingDescription ? location.tournaments[0].address.disambiguatingDescription + ' ' : ''}${location.tournaments[0].address.streetAddress}, ${location.tournaments[0].address.postalCode} ${location.tournaments[0].address.addressLocality}`
                                : 'Adresse non disponible',
                            location.tournaments.map(t => t.rules?.url || (t.affiche ? t.affiche : '')).join(' | '),
                            postalCode,
                            formatCityName(location.tournaments[0].address.addressLocality),
                            getRegionFromPostalCode(postalCode.split(' | ')[0]),
                            location.count,
                            location.count > 1 ? location.count.toString() : '',
                            location.count > 1 ? 1.5 : 1
                        ]
                    })
                };

                const pastTournamentsDataset = {
                    fields: [
                        { name: 'latitude', type: 'real' },
                        { name: 'longitude', type: 'real' },
                        { name: 'Nom du tournoi', type: 'string' },
                        { name: 'Type de tournoi', type: 'string' },
                        { name: 'Club organisateur', type: 'string' },
                        { name: 'Dotation totale (€)', type: 'real', analyzerType: 'INT' },
                        { name: 'Date(s)', type: 'date' },
                        { name: 'Adresse', type: 'string' },
                        { name: 'Règlement', type: 'string' },
                        { name: 'Code postal', type: 'string' },
                        { name: 'Ville', type: 'string' },
                        { name: 'Région', type: 'string' },
                        { name: 'count', type: 'integer' },
                        { name: 'count_display', type: 'string' },
                        { name: 'size_multiplier', type: 'real' }
                    ],
                    rows: Object.values(
                        allPastTournamentsForMap.reduce((acc, t) => {
                            const key = `${t.address.latitude},${t.address.longitude}`;
                            if (!acc[key]) {
                                acc[key] = {
                                    latitude: t.address.latitude,
                                    longitude: t.address.longitude,
                                    tournaments: [],
                                    count: 0
                                };
                            }
                            acc[key].tournaments.push(t);
                            acc[key].count++;
                            return acc;
                        }, {} as Record<string, any>)
                    ).map(location => {
                        console.log('Past Location data:', location);
                        const postalCode = Array.from(new Set<string>(location.tournaments.map(t => formatPostcode(t.address.postalCode)))).join(' | ');
                        return [
                            location.latitude,
                            location.longitude,
                            location.tournaments.map(t => t.name).join(' | '),
                            Array.from(new Set<string>(location.tournaments.map(t => t.type))).join(' | '),
                            Array.from(new Set<string>(location.tournaments.map(t => `${t.club.name} (${t.club.identifier})`))).join(' | '),
                            location.tournaments.map(t =>
                            (typeof t.endowment === 'number' && t.endowment > 0
                                ? Math.floor(t.endowment / 100)
                                : (t.tables?.reduce((sum, table) => sum + (table.endowment || 0), 0) || 0) / 100)
                            ).join(' | '),
                            location.tournaments.map(t =>
                                `${formatDateDDMMYYYY(t.startDate)}${t.startDate !== t.endDate ? ` au ${formatDateDDMMYYYY(t.endDate)}` : ''}`
                            ).join(' | '),
                            location.tournaments[0].address.streetAddress
                                ? `${location.tournaments[0].address.disambiguatingDescription ? location.tournaments[0].address.disambiguatingDescription + ' ' : ''}${location.tournaments[0].address.streetAddress}, ${location.tournaments[0].address.postalCode} ${location.tournaments[0].address.addressLocality}`
                                : 'Adresse non disponible',
                            location.tournaments.map(t => t.rules?.url || 'Pas de règlement')
                                .map(url => location.count > 1 ? url.replace('https://', '') : url)
                                .join(' | '),
                            postalCode,
                            formatCityName(location.tournaments[0].address.addressLocality),
                            getRegionFromPostalCode(postalCode.split(' | ')[0]),
                            location.count,
                            location.count > 1 ? location.count.toString() : '',
                            location.count > 1 ? 1.5 : 1
                        ]
                    })
                };

                store.dispatch(
                    addDataToMap({
                        datasets: [
                            {
                                info: {
                                    label: 'France',
                                    id: 'france'
                                },
                                data: {
                                    fields: [
                                        { name: '_geojson', type: 'geojson' }
                                    ],
                                    rows: [[franceBorders]]
                                }
                            },
                            {
                                info: {
                                    id: 'tournoi'
                                },
                                data: tournamentsDataset
                            },
                        ],
                        options: {
                            centerMap: false,
                            readOnly: false
                        },
                        config: enhancedMapConfig
                    })
                );

                if (pastTournaments?.length > 0) {
                    store.dispatch(
                        addDataToMap({
                            datasets: [
                                {
                                    info: {
                                        id: 'tournoi_passe'
                                    },
                                    data: pastTournamentsDataset
                                }
                            ],
                            options: {
                                centerMap: false,
                                readOnly: false
                            }
                        })
                    );

                    // Manually update the tooltip configuration
                    const currentState = (store.getState().keplerGl as Record<string, any>).paris as {
                        visState: Record<string, any>
                    };
                    const updatedConfig = {
                        ...currentState.visState,
                        interactionConfig: {
                            ...currentState.visState.interactionConfig,
                            tooltip: {
                                ...currentState.visState.interactionConfig.tooltip,
                                fieldsToShow: {
                                    ...currentState.visState.interactionConfig.tooltip.fieldsToShow,
                                    tournoi_passe: [
                                        { name: 'Nom du tournoi', format: null },
                                        { name: 'Type de tournoi', format: null },
                                        { name: 'Club organisateur', format: null },
                                        { name: 'Dotation totale (€)', format: null },
                                        { name: 'Date(s)', format: null },
                                        { name: 'Adresse', format: null },
                                        { name: 'Règlement', format: null }
                                    ]
                                }
                            }
                        }
                    };

                    store.dispatch({
                        type: 'keplerGl/LAYER_CONFIG_CHANGE',
                        payload: {
                            config: updatedConfig
                        }
                    });
                }
            }
        }
    }, [tournaments, pastTournaments]);

    useEffect(() => {
        const hideFranceLayer = () => {
            const franceLayers = document.querySelectorAll(
                '.sortable-layer-item input[id="france_borders:input-layer-label"]'
            );

            franceLayers.forEach(input => {
                const layerItem = input.closest('.sortable-layer-item');
                if (layerItem) {
                    (layerItem as HTMLElement).style.display = 'none';
                }
            });
        };

        // Run immediately and also on potential dynamic content loads
        hideFranceLayer();

        // Optional: Add a MutationObserver for dynamic content
        const observer = new MutationObserver(hideFranceLayer);
        observer.observe(document.body, {
            childList: true,
            subtree: true
        });

        // Cleanup
        return () => {
            observer.disconnect();
        };
    }, []);

    // Spinner component with enhanced visibility and debugging
    const Spinner = () => {
        console.error('SPINNER: Rendering spinner component');
        return (
            <div
                id="loading-spinner"
                style={{
                    position: 'fixed',
                    top: 0,
                    left: 0,
                    width: '100vw',
                    height: '100vh',
                    backgroundColor: '#0E0E0E',
                    display: 'flex',
                    justifyContent: 'center',
                    alignItems: 'center',
                    zIndex: 9999
                }}
            >
                <div
                    style={{
                        width: '100px',
                        height: '100px',
                        border: '10px solid #1FBAD6', // Base turquoise
                        borderTop: '10px solid rgba(31, 186, 214, 0.3)', // Lighter turquoise for spinning effect
                        borderRight: '10px solid rgba(31, 186, 214, 0.6)', // Slightly darker turquoise
                        borderBottom: '10px solid rgba(31, 186, 214, 0.8)', // Even darker turquoise
                        borderRadius: '50%',
                        animation: 'spin 1s linear infinite'
                    }}
                />
                <style>{`
            @keyframes spin {
              0% { transform: rotate(0deg); }
              100% { transform: rotate(360deg); }
            }
          `}</style>
            </div>
        );
    };

    // Debug logging for loading state
    // useEffect(() => {
    //     console.error(`LOADING STATE CHANGED: ${isLoading}`);

    //     // Add a global window variable for debugging
    //     (window as any).loadingState = isLoading;
    // }, [isLoading]);

    // Error boundary component
    class ErrorBoundary extends React.Component<{ children: React.ReactNode }, { hasError: boolean }> {
        constructor(props) {
            super(props);
            this.state = { hasError: false };
        }

        static getDerivedStateFromError(error) {
            return { hasError: true };
        }

        componentDidCatch(error, errorInfo) {
            console.error("Uncaught error:", error, errorInfo);
        }

        render() {
            if (this.state.hasError) {
                return <h1>Something went wrong. Please refresh the page.</h1>;
            }

            return this.props.children;
        }
    }

    // Wrap the entire component with error boundary
    return (
        <ErrorBoundary>
            {isLoading ? (
                <>
                    <Spinner />
                    <div style={{
                        position: 'fixed',
                        top: '50%',
                        left: '50%',
                        transform: 'translate(-50%, -50%)',
                        zIndex: 10000,
                        color: 'red',
                        fontSize: '24px'
                    }}>
                    </div>
                </>
            ) : (
                <div style={{ position: 'absolute', width: '100%', height: '100%' }}>
                    <div aria-hidden="true" style={{
                        position: 'absolute',
                        width: '1px',
                        height: '1px',
                        padding: '0',
                        margin: '-1px',
                        overflow: 'hidden',
                        clip: 'rect(0, 0, 0, 0)',
                        whiteSpace: 'nowrap',
                        border: '0'
                    }}>
                        <h1>Carte Interactive des Tournois de Tennis de Table en France</h1>
                        <p>
                            Bienvenue sur la carte interactive des tournois de tennis de table en France.
                            Trouvez facilement les tournois FFTT près de chez vous grâce à notre carte interactive.
                            Visualisez tous les tournois homologués par la Fédération Française de Tennis de Table.
                        </p>
                        <h2>Fonctionnalités de la Carte des Tournois FFTT</h2>
                        <ul>
                            <li>Visualisation de tous les tournois FFTT sur une carte de France interactive</li>
                            <li>Filtrage par date des tournois de ping-pong</li>
                            <li>Recherche par ville, département et code postal</li>
                            <li>Filtrage par montant de dotation et type de tournoi</li>
                            <li>Accès direct aux règlements des tournois homologués</li>
                            <li>Informations détaillées : club organisateur, dates, adresse, tables</li>
                            <li>Mise à jour en temps réel des compétitions de tennis de table</li>
                        </ul>
                        <h2>Prochains Tournois de Tennis de Table en France</h2>
                        <p>
                            Découvrez les prochains tournois de tennis de table homologués par la FFTT.
                            Notre carte est mise à jour en temps réel avec les dernières informations des clubs.
                            Filtrez par région, date ou type de compétition pour trouver le tournoi qui vous convient.
                        </p>
                        <h2>Compétitions de Tennis de Table par Région</h2>
                        <p>
                            Explorez les tournois de ping-pong dans toute la France :
                            Île-de-France, Auvergne-Rhône-Alpes, Nouvelle-Aquitaine, Occitanie,
                            Hauts-de-France, Grand Est, Provence-Alpes-Côte d'Azur,
                            Normandie, Bretagne, Pays de la Loire, Bourgogne-Franche-Comté,
                            Centre-Val de Loire, Corse et départements d'Outre-mer.
                        </p>
                        <h2>Informations sur les Tournois</h2>
                        <p>
                            Pour chaque tournoi de tennis de table, retrouvez :
                        </p>
                        <ul>
                            <li>Le nom et le type du tournoi FFTT</li>
                            <li>Les dates de début et de fin de la compétition</li>
                            <li>Le montant de la dotation et des récompenses</li>
                            <li>L'adresse complète du gymnase ou de la salle</li>
                            <li>Le club organisateur et son numéro d'affiliation</li>
                            <li>Le règlement officiel du tournoi homologué</li>
                            <li>Le nombre de tables disponibles</li>
                        </ul>
                        <h2>À Propos de la Carte des Tournois</h2>
                        <p>
                            Service gratuit de visualisation des tournois de tennis de table en France.
                            Données officielles issues de la FFTT, mises à jour automatiquement.
                            Interface intuitive pour trouver rapidement les compétitions près de chez vous.
                        </p>
                    </div>
                    <KeplerGl
                        id="paris"
                        width={window.innerWidth}
                        height={window.innerHeight}
                        mapboxApiAccessToken={MAPBOX_TOKEN}
                        localeMessages={{ en: fr }}
                    />
                    <div style={{
                        position: 'absolute',
                        bottom: 5,
                        background: '#242730',
                        height: 20,
                        width: 300,
                        left: 20,
                        color: 'white',
                        fontSize: 10,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                    }}>Données mises à jour en temps réel. | <a style={{ color: 'white', marginLeft: '5px' }} href="/cookies">Cookies et vie privée</a></div>
                </div>
            )}
        </ErrorBoundary>
    );
};