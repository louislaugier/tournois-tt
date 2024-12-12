import React, { useEffect, useState } from 'react';
import { Provider } from 'react-redux';
import { addDataToMap } from '@kepler.gl/actions';
import '@kepler.gl/styles';
import { store } from './lib/store';
import { Map } from './components/Map';
import { DEFAULT_MAP_CONFIG } from './lib/map/config';
import { TournamentQueryBuilder } from './lib/api/tournaments';
import { Tournament } from './lib/api/types';

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
        console.log('Raw tournament data (first tournament):', JSON.stringify(tournamentData?.[0], null, 2));
        console.log('Tournament data length:', tournamentData?.length ?? 0);
        setTournaments(tournamentData || []);
      } catch (error) {
        console.error('Failed to load tournaments:', error);
      }
    };

    loadTournaments();
  }, []);

  useEffect(() => {
    if (tournaments.length > 0) {
      const tournamentsWithCoordinates = tournaments.filter(
        t => t.address?.latitude && t.address?.longitude
      );
      const tournamentsWithoutCoordinates = tournaments.filter(
        t => !t.address?.latitude || !t.address?.longitude
      );
      
      console.log('Tournaments with coordinates:', tournamentsWithCoordinates.length);
      console.log('Tournaments without coordinates:', tournamentsWithoutCoordinates.length);
      console.log('Tournaments without coordinates details:', 
        tournamentsWithoutCoordinates.map(t => ({
          name: t.name,
          address: t.address,
          hasLatitude: !!t.address?.latitude,
          hasLongitude: !!t.address?.longitude
        }))
      );
      
      // Combine tournaments with and without coordinates
      const allTournamentsForMap = [
        ...tournamentsWithCoordinates,
        ...tournamentsWithoutCoordinates.map(t => ({
          ...t,
          address: {
            ...t.address,
            latitude: 46.777138, // France center latitude
            longitude: 2.804568, // France center longitude
            approximate: true
          }
        }))
      ];
      
      if (allTournamentsForMap.length > 0) {
        const mapData = {
          fields: [
            { name: 'Nom du tournoi', type: 'string' },
            { name: 'latitude', type: 'real' },
            { name: 'longitude', type: 'real' },
            { name: 'Type', type: 'string' },
            { name: 'Date de début', type: 'string' },
            { name: 'Date de fin', type: 'string' },
            { name: 'Club', type: 'string' },
            { name: 'Adresse', type: 'string' },
            { name: 'Règlement', type: 'string' },
            { name: 'Dotation totale', type: 'string' },
            { name: 'Localisation approximative', type: 'string' },
          ],
          rows: allTournamentsForMap.map(t => {
            const endowmentStr = typeof t.endowment === 'number' && t.endowment > 0 
              ? (t.endowment / 100).toFixed(2) + '€' 
              : '';

            return [
              t.name || '',
              t.address.latitude,
              t.address.longitude,
              t.type || '',
              new Date(t.startDate).toLocaleDateString('fr-FR', {
                weekday: 'long',
                day: 'numeric',
                month: 'long',
                year: 'numeric'
              }),
              new Date(t.endDate).toLocaleDateString('fr-FR', {
                weekday: 'long',
                day: 'numeric',
                month: 'long',
                year: 'numeric'
              }),
              t.club.name ? `${t.club.name}${t.club.identifier ? ` (${t.club.identifier})` : ''}` : '',
              t.address.streetAddress ? `${t.address.streetAddress}, ${t.address.postalCode} ${t.address.addressLocality}` : '',
              t.rules?.url || '',
              endowmentStr,
              'oui', // Always mark as approximate for tournaments without coordinates
            ];
          })
        };

        console.log('Final mapData:', mapData);

        store.dispatch(
          addDataToMap({
            datasets: [{
              info: {
                label: 'Tournois FFTT',
                id: 'tournament_data'
              },
              data: mapData
            }],
            options: {
              centerMap: false,
              readOnly: false
            },
            config: DEFAULT_MAP_CONFIG
          })
        );
      }
    }
  }, [tournaments]);

  return (
    <Provider store={store}>
      <div style={{ position: 'absolute', width: '100%', height: '100%' }}>
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