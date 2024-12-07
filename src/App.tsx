import React from 'react';
import { Provider } from 'react-redux';
import { addDataToMap } from '@kepler.gl/actions';
import '@kepler.gl/styles';
import { store } from './lib/store';
import { Map } from './components/Map';
import { INITIAL_MAP_DATA } from './lib/map-config/constants';
import { DEFAULT_MAP_CONFIG } from './lib/map-config/config';

const App = () => {
  React.useEffect(() => {
    store.dispatch(
      addDataToMap({
        datasets: [{
          info: {
            label: 'France Locations',
            id: 'france_data'
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