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
      console.log('Tournaments with coordinates:', tournamentsWithCoordinates.length);
      
      if (tournamentsWithCoordinates.length > 0) {
        const mapData = {
          fields: [
            { name: 'Tournoi', type: 'string' },
            { name: 'latitude', type: 'real' },
            { name: 'longitude', type: 'real' },
            { name: 'Date de début', type: 'string' },
            { name: 'Date de fin', type: 'string' },
            { name: 'Club', type: 'string' },
            { name: 'Adresse', type: 'string' },
            { name: 'approximate', type: 'boolean' },
          ],
          rows: tournamentsWithCoordinates.map(t => [
            t.name,
            t.address.latitude,
            t.address.longitude,
            t.startDate,
            t.endDate,
            `${t.club.name}${t.club.identifier ? ` (${t.club.identifier})` : ''}`,
            `${t.address.streetAddress}, ${t.address.postalCode} ${t.address.addressLocality}`,
            t.address.approximate || false,
          ])
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