import { useEffect } from 'react';
import { Provider } from 'react-redux';
import { addDataToMap } from '@kepler.gl/actions';
import '@kepler.gl/styles';
import { store } from './lib/store';
import { Map } from './components/Map';
import { INITIAL_MAP_DATA } from './lib/map/constants';
import { DEFAULT_MAP_CONFIG } from './lib/map/config';
import { TournamentQueryBuilder } from './lib/api/tournaments';

const App = () => {
  useEffect(() => {
    const loadTournaments = async () => {
      try {
        const today = new Date();
        const query = new TournamentQueryBuilder()
          .startDateRange(today)
          .orderByStartDate('asc')
          .itemsPerPage(999999);

        await query.executeAndLogAll();
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