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

const CustomMapPopover: React.FC<any> = ({ data }) => {
  // If no data or empty, return null
  if (!data || data.length === 0) return null;

  // For single point, show detailed tournament info
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

  // For multiple points, show summary
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
        {data.map((point: any, index: any) => (
          <li key={index}>{point.Tournoi}</li>
        ))}
      </ul>
    </div>
  );
};

// Replace the default MapPopoverFactory
const CustomMapPopoverFactory = () => CustomMapPopover;
CustomMapPopoverFactory.deps = MapPopoverFactory.deps;

const App = () => {
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

      // Combine tournaments with and without coordinates
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
                    { name: 'Tournoi', format: null },
                    { name: 'Type', format: null },
                    { name: 'Club', format: null },
                    { name: 'Dotation', format: null },
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
            layers: [{
              id: 'tournament_points',
              type: 'point',
              config: {
                dataId: 'tournoi',
                label: 'Tournois',
                color: [64, 224, 208] as [number, number, number],
                columns: {
                  lat: 'latitude',
                  lng: 'longitude'
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
            pitch: 0, // Reset pitch to default
            bearing: 0,
            dragRotate: false // Disable rotation
          }
        };

        const tournamentsDataset = {
          fields: [
            { name: 'latitude', type: 'real' },
            { name: 'longitude', type: 'real' },
            { name: 'Tournoi', type: 'string' },
            { name: 'Type', type: 'string' },
            { name: 'Club', type: 'string' },
            { name: 'Dotation', type: 'string' },
            { name: 'Dates', type: 'string' },
            { name: 'Adresse', type: 'string' },
            { name: 'Règlement', type: 'string' }
          ],
          rows: allTournamentsForMap.map(t => [
            t.address.latitude,
            t.address.longitude,
            t.name || '',
            t.type || '',
            t.club.name ? `${t.club.name}${t.club.identifier ? ` (${t.club.identifier})` : ''}` : '',
            typeof t.endowment === 'number' && t.endowment > 0 
              ? `${Math.floor(t.endowment / 100)}€` 
              : '?',
            `${new Date(t.startDate).toLocaleDateString('fr-FR')} - ${new Date(t.endDate).toLocaleDateString('fr-FR')}`,
            t.address.streetAddress
              ? `${t.address.disambiguatingDescription ? t.address.disambiguatingDescription + ' ' : ''}${t.address.streetAddress}, ${t.address.postalCode} ${t.address.addressLocality}`
              : '',
            t.rules?.url || ''
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