import React, { useEffect, useState } from 'react';
import { Provider } from 'react-redux';
import { addDataToMap } from '@kepler.gl/actions';
import '@kepler.gl/styles';
import { store } from './lib/store';
import { Map } from './components/Map';
import { DEFAULT_MAP_CONFIG } from './lib/map/config';
import { TournamentQueryBuilder } from './lib/api/tournaments';
import { Tournament } from './lib/api/types';
import { MapPopoverFactory } from '@kepler.gl/components';

const formatDateDDMMYYYY = (date: Date | string) => {
  const d = new Date(date);
  return d.toLocaleDateString('fr-FR', { 
    day: '2-digit', 
    month: '2-digit', 
    year: 'numeric' 
  });
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
        <p><strong>Dates:</strong> {point.Dates}</p>
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
          <li key={index}>{point.Tournoi}</li>
        ))}
      </ul>
    </div>
  );
};

const CustomMapPopoverFactory = () => CustomMapPopover;
CustomMapPopoverFactory.deps = MapPopoverFactory.deps;

const App: React.FC = () => {
  const [tournaments, setTournaments] = useState<Tournament[]>([]);

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
      ];

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
                    { name: 'Dotation totale', format: null },
                    { name: 'Dates', format: null },
                    { name: 'Adresse', format: null },
                    { name: 'Règlement', format: null }
                  ]
                },
                enabled: true,
                compareMode: false,
                compareType: 'absolute'
              }
            },
            filters: [
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
                id: 'date_filter',
                dataId: ['tournoi'],
                name: ['Date de début'],
                type: 'timeRange',
                value: [
                  Math.min(...tournaments.map(t => new Date(t.startDate).getTime())),
                  Math.max(...tournaments.map(t => new Date(t.startDate).getTime()))
                ],
                enlarged: true,
                plotType: 'histogram',
                layerId: undefined,
                field: {
                  type: 'timestamp',
                  name: 'Date de début'
                }
              },
              {
                id: 'endowment_filter',
                dataId: ['tournoi'],
                name: ['Dotation totale'],
                type: 'range',
                value: [0, 10000],
                enlarged: false,
                plotType: 'histogram',
                layerId: undefined,
                field: {
                  type: 'real',
                  name: 'Dotation totale'
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
              }
            ],
            layers: [{
              id: 'tournament_points',
              type: 'point',
              config: {
                dataId: 'tournoi',
                label: 'Tournoi',
                color: [64, 224, 208] as [number, number, number],
                columns: {
                  lat: 'latitude',
                  lng: 'longitude',
                  altitude: 'Dotation'
                },
                isVisible: true,
                visConfig: {
                  radius: 13,
                  fillColor: [64, 224, 208] as [number, number, number],
                  color: [64, 224, 208] as [number, number, number],
                  opacity: 0.8
                }
              },
              visualChannels: {}
            }]
          },
          mapState: {
            ...DEFAULT_MAP_CONFIG.mapState,
            pitch: 0,
            bearing: 0,
            dragRotate: false
          }
        };

        const tournamentsDataset = {
          fields: [
            { name: 'latitude', type: 'real' },
            { name: 'longitude', type: 'real' },
            { name: 'Nom du tournoi', type: 'string' },
            { name: 'Type de tournoi', type: 'string' },
            { name: 'Club organisateur', type: 'string' },
            { name: 'Dotation totale', type: 'real' },
            { name: 'Dates', type: 'date' },
            { name: 'Date de début', type: 'date' },
            { name: 'Adresse', type: 'string' },
            { name: 'Règlement', type: 'string' }
          ],
          rows: allTournamentsForMap.map(t => [
            t.address.latitude,
            t.address.longitude,
            t.name || 'Tournoi sans nom',
            t.type || 'Non spécifié',
            t.club.name ? `${t.club.name}${t.club.identifier ? ` (${t.club.identifier})` : ''}` : 'Club non spécifié',
            typeof t.endowment === 'number' && t.endowment > 0 
              ? Math.floor(t.endowment / 100)
              : 0,
            `${formatDateDDMMYYYY(t.startDate)} - ${formatDateDDMMYYYY(t.endDate)}`,
            new Date(t.startDate).getTime(),
            t.address.streetAddress
              ? `${t.address.disambiguatingDescription ? t.address.disambiguatingDescription + ' ' : ''}${t.address.streetAddress}, ${t.address.postalCode} ${t.address.addressLocality}`
              : 'Adresse non disponible',
            t.rules?.url || 'Pas de règlement'
          ])
        };

        store.dispatch(
          addDataToMap({
            datasets: [{
              info: {
                label: 'Tournois FFTT',
                id: 'tournoi'
              },
              data: tournamentsDataset
            }],
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
    <Provider store={store}>
      <div style={{ position: 'absolute', width: '100%', height: '100%', opacity: !!tournaments.length ? 1 : 0 }}>
        <Map
          id="paris"
          width={window.innerWidth}
          height={window.innerHeight}
        />
      </div>
    </Provider>
  );
};

export default App;
