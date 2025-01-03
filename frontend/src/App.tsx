import React, { useEffect, useState } from 'react';
import { Helmet } from 'react-helmet';
import { Provider } from 'react-redux';
import { BrowserRouter as Router, Routes, Route, useLocation } from 'react-router-dom';
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

  useEffect(() => {
    initializeSidebarCustomizer();
    initializeDateFormatter();
    initializeDOMCleaner();
  }, []);

  useEffect(() => {
    const loadTournaments = async () => {
      try {
        const today = new Date();
        const query = new TournamentQueryBuilder()
          .startDateRange(today)
          .orderByStartDate('asc')
          .itemsPerPage(999999);

        const tournamentData = await query.executeAndLogAll();
        setTournaments(tournamentData || []);
      } catch (error) {
        console.error('Failed to load tournaments:', error);
      }
    };

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

      if (allTournamentsForMap.length > 0) {
        const enhancedMapConfig = {
          ...DEFAULT_MAP_CONFIG,
          visState: {
            ...DEFAULT_MAP_CONFIG.visState,
            interactionConfig: {
              ...DEFAULT_MAP_CONFIG.visState.interactionConfig,
              tooltip: {
                fieldsToShow: {
                  tournoi: [
                    { name: 'Nom du tournoi', format: null },
                    { name: 'Type de tournoi', format: null },
                    { name: 'Club organisateur', format: null },
                    { name: 'Dotation totale (€)', format: null },
                    { name: 'Date(s)', format: null },
                    { name: 'Adresse', format: null },
                    { name: 'Règlement', format: null }
                  ],
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
                name: ['Dates de début des tournois'],
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
                  name: 'Dates de début des tournois'
                }
              },
              {
                id: 'endowment_filter',
                dataId: ['tournoi'],
                name: ['Dotation totale (€)'],
                type: 'range',
                value: null,
                enlarged: false,
                plotType: 'histogram',
                layerId: undefined,
                field: {
                  type: 'real',
                  name: 'Dotation totale (€)'
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
                  label: 'Tournoi',
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
                    size: 14,
                    offset: [0, 0] as [number, number],
                    anchor: 'middle',
                    alignment: 'center',
                    background: false,
                    outlineWidth: 0,
                    outlineColor: [0, 0, 0, 0] as [number, number, number, number],
                    backgroundColor: [0, 0, 0, 0] as [number, number, number, number]
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
                id: 'france_borders',
                type: 'geojson',
                config: {
                  dataId: 'france',
                  label: 'France',
                  color: [255, 255, 255] as [number, number, number],
                  columns: {
                    geojson: '_geojson'
                  } as { [key: string]: string },
                  isVisible: true,
                  visConfig: {
                    opacity: 1,
                    thickness: 1,
                    strokeColor: [255, 255, 255] as [number, number, number],
                    filled: false,
                    stroked: true,
                    enable3d: false,
                    pickable: false,
                    fixedRadius: false,
                    radiusRange: [0, 0],
                    clusterRadius: 0
                  },
                  textLabel: {
                    field: { name: '', type: 'string' },
                    color: [255, 255, 255] as [number, number, number],
                    size: 12,
                    offset: [0, 0] as [number, number],
                    anchor: 'start',
                    alignment: 'center',
                    outlineWidth: 0,
                    outlineColor: [0, 0, 0, 0] as [number, number, number, number],
                    background: false,
                    backgroundColor: [0, 0, 0, 0] as [number, number, number, number]
                  },
                  isConfigActive: true,
                  colorUI: {
                    color: {
                      type: 'none',
                      defaultValue: null
                    }
                  },
                  interaction: {
                    tooltip: {
                      enabled: false,
                      fieldsToShow: {}
                    },
                    brush: {
                      enabled: false
                    },
                    coordinate: {
                      enabled: false
                    },
                    clicked: false,
                    hovered: false
                  }
                },
                visualChannels: {
                  colorField: { name: '', type: 'string' },
                  colorScale: 'quantile',
                  sizeField: { name: '', type: 'string' },
                  sizeScale: 'linear',
                  strokeColorField: { name: '', type: 'string' },
                  strokeColorScale: 'quantile'
                }
              }
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

        const tournamentsDataset = {
          fields: [
            { name: 'latitude', type: 'real' },
            { name: 'longitude', type: 'real' },
            { name: 'Nom du tournoi', type: 'string' },
            { name: 'Type de tournoi', type: 'string' },
            { name: 'Club organisateur', type: 'string' },
            { name: 'Dotation totale (€)', type: 'real', analyzerType: 'INT' },
            { name: 'Date(s)', type: 'date' },
            { name: 'Dates de début des tournois', type: 'date' },
            { name: 'Adresse', type: 'string' },
            { name: 'Règlement', type: 'string' },
            { name: 'Code postal', type: 'string' },
            { name: 'Ville', type: 'string' },
            { name: 'count', type: 'integer' },
            { name: 'count_display', type: 'string' },
            { name: 'size_multiplier', type: 'real' }
          ],
          rows: Object.values(
            allTournamentsForMap.reduce((acc, t) => {
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
          ).map(location => [
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
            location.tournaments.map(t => t.rules?.url || 'Pas encore de règlement').join(' | '),
            Array.from(new Set<string>(location.tournaments.map(t => formatPostcode(t.address.postalCode)))).join(' | '),
            formatCityName(location.tournaments[0].address.addressLocality),
            location.count,
            location.count > 1 ? location.count.toString() : '',
            location.count > 1 ? 1.5 : 1
          ])
        };

        store.dispatch(
          addDataToMap({
            datasets: [
              {
                info: {
                  label: 'Tournois FFTT',
                  id: 'tournoi'
                },
                data: tournamentsDataset
              },
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
              }
            ],
            options: {
              centerMap: false,
              readOnly: false
            },
            config: enhancedMapConfig
          })
        );
      }
    }
  }, [tournaments]);

  return (
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
