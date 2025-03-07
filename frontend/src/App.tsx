import React, { useEffect, useState } from 'react';
import { Helmet } from 'react-helmet';
import { Provider } from 'react-redux';
import { BrowserRouter as Router, Routes, Route, useLocation, Navigate } from 'react-router-dom';
import { addDataToMap } from '@kepler.gl/actions';
import '@kepler.gl/styles';
import { initializeDateFormatter } from './lib/dateFormatter';
import { initializeSidebarCustomizer } from './lib/sidebarCustomizer';
import { initializeDOMCleaner } from './lib/domCleaner';
import Cookies from './pages/Cookies';

// Disable error overlay in production
if (process.env.NODE_ENV === 'production') {
  console.error = () => { };
  window.addEventListener('error', (e) => {
    e.stopPropagation();
    e.preventDefault();
  });
  window.addEventListener('unhandledrejection', (e) => {
    e.stopPropagation();
    e.preventDefault();
  });
}

import { store } from './lib/store';
import { DEFAULT_MAP_CONFIG } from './lib/map/config';
import { TournamentQueryBuilder } from './lib/api/tournaments';
import { Tournament } from './lib/api/types';
import { KeplerGl, MapPopoverFactory } from '@kepler.gl/components';
import franceBordersRaw from './assets/metropole-et-outre-mer.json';
import { MAPBOX_TOKEN } from './lib/map/constants';
import fr from './locales/fr';

const franceBorders = typeof franceBordersRaw === 'string' ? JSON.parse(franceBordersRaw) : franceBordersRaw;

const formatDateDDMMYYYY = (date: Date | string) => {
  const d = new Date(date);
  const day = d.getDate().toString().padStart(2, '0');
  const month = (d.getMonth() + 1).toString().padStart(2, '0');
  const year = d.getFullYear();
  return `${day}/${month}/${year}`;
};

const formatPostcode = (postcode: string | undefined): string => {
  if (!postcode) return '';  // Return empty string for undefined/null
  return postcode.toString();
};

const formatCityName = (city: string): string => {
  const upperCity = city?.toUpperCase() || '';

  // Special case for VILLENEUVE D ASCQ
  if (upperCity === 'VILLENEUVE D ASCQ') {
    return "VILLENEUVE D'ASCQ";
  } else if (upperCity === '51000 - CHALONS EN CHAMPAGNE') {
    return "CHALONS-EN-CHAMPAGNE";
  }

  // Replace STE and ST with full words
  const cityWithFullWords = upperCity
    .replace(/\bSTE\b/g, 'SAINTE')
    .replace(/\bST\b/g, 'SAINT');

  // Split the string into words
  const words = cityWithFullWords.split(' ');
  const result: string[] = [];

  for (let i = 0; i < words.length; i++) {
    const currentWord = words[i];
    const nextWord = words[i + 1];

    // Add current word
    result.push(currentWord);

    // If there's a next word and current word is not LA/LE/LES, add hyphen
    if (nextWord && !['LA', 'LE', 'LES'].includes(currentWord)) {
      result.push('-');
    } else if (nextWord) {
      // If it is LA/LE/LES, add space
      result.push(' ');
    }
  }

  return result.join('');
};

const getRegionFromPostalCode = (postalCode: string): string => {
  const prefix = postalCode.substring(0, 2);
  const prefixNum = parseInt(prefix, 10);

  // Handle special cases for overseas departments
  if (['971', '972', '973', '974', '976'].includes(postalCode.substring(0, 3))) {
    return 'Outre-Mer';
  }

  // Map department prefixes to regions
  if ([1, 3, 7, 15, 26, 38, 42, 43, 63, 69, 73, 74].includes(prefixNum)) {
    return 'Auvergne-Rhône-Alpes';
  } else if ([21, 25, 39, 58, 70, 71, 89, 90].includes(prefixNum)) {
    return 'Bourgogne-Franche-Comté';
  } else if ([22, 29, 35, 56].includes(prefixNum)) {
    return 'Bretagne';
  } else if ([18, 28, 36, 37, 41, 45].includes(prefixNum)) {
    return 'Centre-Val de Loire';
  } else if ([20].includes(prefixNum)) {
    return 'Corse';
  } else if ([8, 10, 51, 52, 54, 55, 57, 67, 68, 88].includes(prefixNum)) {
    return 'Grand Est';
  } else if ([2, 59, 60, 62, 80].includes(prefixNum)) {
    return 'Hauts-de-France';
  } else if ([75, 77, 78, 91, 92, 93, 94, 95].includes(prefixNum)) {
    return 'Île-de-France';
  } else if ([14, 27, 50, 61, 76].includes(prefixNum)) {
    return 'Normandie';
  } else if ([16, 17, 19, 23, 24, 33, 40, 47, 64, 79, 86, 87].includes(prefixNum)) {
    return 'Nouvelle-Aquitaine';
  } else if ([9, 11, 12, 30, 31, 32, 34, 46, 48, 65, 66, 81, 82].includes(prefixNum)) {
    return 'Occitanie';
  } else if ([44, 49, 53, 72, 85].includes(prefixNum)) {
    return 'Pays de la Loire';
  } else if ([4, 5, 6, 13, 83, 84].includes(prefixNum)) {
    return 'Provence-Alpes-Côte d\'Azur';
  }

  return 'Région inconnue';
};

const CustomMapPopover: React.FC<any> = ({ data }) => {
  if (!data || data.length === 0) return null;

  if (data.length === 1) {
    const point = data[0];
    return (
      <div style={{
        backgroundColor: 'white',
        padding: '10px',
        borderRadius: '5px',
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
      }}>
        <h4>{point.Tournoi}</h4>
        <p><strong>Type:</strong> {point.Type}</p>
        <p><strong>Club:</strong> {point.Club}</p>
        <p><strong>Dotation:</strong> {point.Dotation}</p>
        <p><strong>Date(s):</strong> {point.Dates}</p>
        <p><strong>Adresse:</strong> {point.Adresse}</p>
        {point.Règlement && (
          <p>
            <strong>Règlement:</strong>
            <a href={point.Règlement} target="_blank" rel="noopener noreferrer">
              Voir le règlement
            </a>
          </p>
        )}
      </div>
    );
  }

  return (
    <div style={{
      backgroundColor: 'white',
      padding: '10px',
      borderRadius: '5px',
      boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
    }}>
      <h4>Plusieurs Tournois</h4>
      <p><strong>Nombre de Tournois:</strong> {data.length}</p>
      <p><strong>Tournois:</strong></p>
      <ul>
        {data.map((point: any, index: number) => (
          <li key={index}>
            <strong>{point.Tournoi}</strong>
            <br />
            <span>Du {point.Dates}</span>
          </li>
        ))}
      </ul>
    </div>
  );
};

const CustomMapPopoverFactory = () => CustomMapPopover;
CustomMapPopoverFactory.deps = MapPopoverFactory.deps;

const MapView: React.FC = () => {
  const [tournaments, setTournaments] = useState<Tournament[]>([]);
  const [pastTournaments, setPastTournaments] = useState<Tournament[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  useEffect(() => {
    initializeSidebarCustomizer();
    initializeDateFormatter();
    initializeDOMCleaner();
  }, []);

  useEffect(() => {
    const loadTournaments = async () => {
      console.log('Starting loadTournaments function');
      try {
        console.log('Setting isLoading to true');
        setIsLoading(true);
        const yesterday = new Date();
        yesterday.setDate(yesterday.getDate() - 1);
        yesterday.setHours(0, 0, 0, 0);

        console.log('Creating tournament query');
        const query = new TournamentQueryBuilder()
          .startDateRange(yesterday)
          .orderByStartDate('asc')
          .itemsPerPage(999999);

        console.log('Executing tournament query');
        const tournamentData = await query.executeAndLogAll();
        console.log('Tournament data fetched:', tournamentData?.length || 'No data');

        // Add mock tournament
        const mockTournaments: Array<Tournament> = [
          {
            affiche: 'https://cdn.pepsup.com/resources/images/ARTICLES/000/184/552/1845523/IMAGE/1845523.jpg?1732633837000',
            '@id': '/tournaments/mock-kb-2025',
            '@type': 'Tournament',
            id: 999999, // Unique mock ID
            identifier: 'MOCK-KB-2025',
            name: 'TOURNOI NATIONAL DU KREMLIN-BICETRE',
            type: 'National B',
            club: {
              '@id': '/clubs/uskb',
              '@type': 'Club',
              id: 69, // Unique mock ID
              name: 'KREMLIN-BICETRE T.T.',
              identifier: '08940975'
            },
            startDate: new Date('2025-06-14').toISOString(),
            endDate: new Date('2025-06-15').toISOString(),
            address: {
              '@id': '/addresses/kb-gymnase',
              '@type': 'PostalAddress',
              id: 999999, // Unique mock ID
              postalCode: '94270',
              streetAddress: '12 bd Chastenet de Géry',
              disambiguatingDescription: 'COSEC Elisabeth et Vincent Purkart',
              addressCountry: 'FR',
              addressRegion: 'Ile-de-France',
              addressLocality: 'Le Kremlin-Bicêtre',
              areaServed: null,
              latitude: 48.807173,
              longitude: 2.35724,
              name: 'COSEC Elisabeth et Vincent Purkart',
              identifier: null,
              openingHours: null,
              main: false,
              approximate: false
            },
            contacts: [], // Empty contacts array
            rules: {
              url: ""
            },
            endowment: 0, // Total endowment in cents
            status: 1, // Assuming 1 means active/upcoming
            organization: undefined, // Optional field
            responses: [], // Optional field
            engagmentSheet: undefined, // Optional field
            decision: undefined, // Optional field
            page: null, // Optional field
            '@permissions': {
              canUpdate: true,
              canDelete: false
            }
          },
          {
            affiche: 'https://www.esvitrytt.fr/kcfinder/upload/files/ES%20VITRY%20TT%20TOURNOI%20NATIONAL%20B.pdf',
            '@id': '/tournaments/mock-vitry-2025',
            '@type': 'Tournament',
            id: 999998, // Unique mock ID
            identifier: 'MOCK-VITRY-2025',
            name: 'ES VITRY TENNIS DE TABLE TOURNOI NATIONAL B',
            type: 'National B',
            club: {
              '@id': '/clubs/esvitry',
              '@type': 'Club',
              id: 68, // Unique mock ID
              name: 'VITRY ES',
              identifier: '08940448'
            },
            startDate: new Date('2025-04-26').toISOString(),
            endDate: new Date('2025-04-27').toISOString(),
            address: {
              '@id': '/addresses/vitry-gymnase',
              '@type': 'PostalAddress',
              id: 999998, // Unique mock ID
              postalCode: '94400',
              streetAddress: '4 Avenue du Colonel Fabien',
              disambiguatingDescription: 'GYMNASE GOSNAT',
              addressCountry: 'FR',
              addressRegion: 'Ile-de-France',
              addressLocality: 'Vitry-sur-Seine',
              areaServed: null,
              latitude: 48.7815592,
              longitude: 2.3802291,
              name: 'GYMNASE GOSNAT',
              identifier: null,
              openingHours: null,
              main: false,
              approximate: false
            },
            contacts: [], // Empty contacts array
            rules: {
              url: ""
            },
            endowment: 131000, // Total endowment in cents
            status: 1, // Assuming 1 means active/upcoming
            organization: undefined, // Optional field
            responses: [], // Optional field
            engagmentSheet: undefined, // Optional field
            decision: undefined, // Optional field
            page: null, // Optional field
            '@permissions': {
              canUpdate: true,
              canDelete: false
            }
          }
        ];

        const normalizeDate = (date) => {
          const d = new Date(date);
          d.setHours(0, 0, 0, 0); // Set the time to midnight
          return d.getTime();
        };
        const isSameDay = (t: Tournament) => normalizeDate(t.startDate) === normalizeDate(mockTournaments[0].startDate);

        let tournamentDataToSet = tournamentData || [];
        mockTournaments.forEach(mockTournament => {
          // Add mock tournament to tournament data only if it's not already in the data (check club identifier & if on the same date)
          const existingTournament = tournamentDataToSet.find(t => 
            t.club.identifier === mockTournament.club.identifier && 
            normalizeDate(t.startDate) === normalizeDate(mockTournament.startDate)
          );
          if (!existingTournament) {
            tournamentDataToSet = [...tournamentDataToSet, mockTournament];
          }
        });
        setTournaments(tournamentDataToSet);

        // Get the current date and determine the latest finished season
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

        const lastCompletedSeasonStartDate = new Date(pastSeasonStartYear, 6, 1); // July 1st of past season start year
        lastCompletedSeasonStartDate.setHours(0, 0, 0, 0);

        const lastCompletedSeasonEndDate = new Date(pastSeasonEndYear, 5, 30); // June 30th of past season end year
        lastCompletedSeasonEndDate.setHours(23, 59, 59, 999);

        const pastTournamentsQuery = new TournamentQueryBuilder()
          .startDateRange(lastCompletedSeasonStartDate, lastCompletedSeasonEndDate)
          .orderByStartDate('asc')
          .itemsPerPage(999999);

        const pastTournamentsData = await pastTournamentsQuery.executeAndLogAll();
        setPastTournaments(pastTournamentsData || []);
      } catch (error) {
        console.error('Failed to load tournaments:', error);
      } finally {
        console.log('Setting isLoading to false');
        setIsLoading(false);
      }
    };

    console.log('Calling loadTournaments');
    loadTournaments();
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

  const now = new Date();
  const currentYear = now.getFullYear();
  const currentMonth = now.getMonth(); // JavaScript months are 0-based

  // Match Go function's logic
  let seasonStartYear, seasonEndYear;
  if (currentMonth >= 6) { // July or later (0-based, so 6 is July)
    // Current season started this year, so last finished season was previous year
    seasonStartYear = currentYear;
    seasonEndYear = currentYear + 1;
  } else {
    // We're in January-June, so current season started previous year
    seasonStartYear = currentYear - 1;
    seasonEndYear = currentYear;
  }

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
            width: '200px', 
            height: '200px', 
            border: '20px solid #1FBAD6', // Turquoise from map
            borderTop: '20px solid #9B59B6', // Purple from map
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
  useEffect(() => {
    console.error(`LOADING STATE CHANGED: ${isLoading}`);
    
    // Add a global window variable for debugging
    (window as any).loadingState = isLoading;
  }, [isLoading]);

  // Error boundary component
  class ErrorBoundary extends React.Component<{children: React.ReactNode}, {hasError: boolean}> {
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

const AppContent: React.FC = () => {
  const location = useLocation();

  useEffect(() => {
    if (location.pathname === '/') {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [location]);

  return (
    <div>
      <Helmet>
        <title>Carte des Tournois de Tennis de Table homologués FFTT</title>
        <meta name="description" content="Découvrez tous les tournois de tennis de table en France sur une carte interactive. Filtrez par date, région, et catégorie. Informations détaillées sur les règlements, dates et lieux des compétitions FFTT." />
        <meta name="keywords" content="tennis de table, tournois, FFTT, ping pong, France, carte interactive, compétitions" />
        <meta property="og:title" content="Carte des Tournois de Tennis de Table en France | FFTT" />
        <meta property="og:description" content="Découvrez tous les tournois de tennis de table en France sur une carte interactive. Filtrez par date, région, et catégorie." />
        <meta property="og:type" content="website" />
        <meta property="og:url" content="https://tournois-tt.fr" />
        <meta property="og:image" content="https://cdn-icons-png.flaticon.com/512/9978/9978844.png" />
        <meta property="og:site_name" content="Carte des Tournois FFTT" />
        <meta name="twitter:card" content="summary_large_image" />
        <meta name="twitter:title" content="Carte des Tournois de Tennis de Table en France" />
        <meta name="twitter:description" content="Découvrez tous les tournois de tennis de table en France sur une carte interactive. Filtrez par date, région, et catégorie." />
        <meta name="twitter:image" content="https://cdn-icons-png.flaticon.com/512/9978/9978844.png" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <meta name="theme-color" content="#242730" />
        <meta name="robots" content="index, follow" />
        <link rel="canonical" href="https://tournois-tt.fr" />
        <script type="application/ld+json">
          {`
            {
              "@context": "https://schema.org",
              "@type": "SportsEvent",
              "name": "Tournois de Tennis de Table en France",
              "description": "Carte interactive des tournois de tennis de table en France avec filtres par date, région et catégorie",
              "sport": "Tennis de Table",
              "location": {
                "@type": "Country",
                "name": "France"
              },
              "organizer": {
                "@type": "Organization",
                "name": "Fédération Française de Tennis de Table",
                "alternateName": "FFTT"
              },
              "eventStatus": "EventScheduled",
              "eventAttendanceMode": "OfflineEventAttendanceMode",
              "offers": {
                "@type": "Offer",
                "availability": "https://schema.org/InStock",
                "price": "0",
                "priceCurrency": "EUR"
              }
            }
          `}
        </script>
      </Helmet>
      <Routes>
        <Route path="/" element={<MapView />} />
        <Route path="/cookies" element={<Cookies />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </div>
  );
};

const App: React.FC = () => {
  return (
    <Provider store={store}>
      <Router>
        <AppContent />
      </Router>
    </Provider>
  );
};

export default App;
