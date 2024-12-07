import React, { useEffect, useState } from 'react';
import { Provider } from 'react-redux';
import { addDataToMap } from '@kepler.gl/actions';
import '@kepler.gl/styles';
import { store } from './lib/store';
import { Map } from './components/Map';
import { INITIAL_MAP_DATA } from './lib/map/constants';
import { DEFAULT_MAP_CONFIG } from './lib/map/config';
import { fetchAllTournaments } from './lib/fftt-api/client';
import type { Tournament } from './lib/fftt-api/types';

const App = () => {
  const [tournaments, setTournaments] = useState<Tournament[]>([]);

  useEffect(() => {
    const loadTournaments = async () => {
      try {
        const tournaments = await fetchAllTournaments({
          'order[startDate]': 'asc',
          'endDate[before]': '2024-12-31T23:59:59'
        });
        setTournaments(tournaments);
        console.log(`Loaded ${tournaments.length} tournaments`);
        tournaments.forEach(t => {
          if (t.address?.streetAddress) {
            console.log('Street Address:', t.address.streetAddress);
          }
        });
      } catch (error) {
        console.error('Failed to load tournaments:', error);
      }
    };

    loadTournaments();
  }, []);

  useEffect(() => {
    store.dispatch(
      addDataToMap({
        datasets: [{
          info: {
            label: 'Tournois FFTT',
            id: 'tournament_data'
          },
          data: INITIAL_MAP_DATA
        }],
        options: {
          centerMap: false,
          readOnly: false
        },
        config: DEFAULT_MAP_CONFIG
      })
    );
  }, []);

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