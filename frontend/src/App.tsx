import React from 'react';
import { Provider } from 'react-redux';
import '@kepler.gl/styles';
import { store } from './lib/store';
import Router from './components/Router';

const App: React.FC = () => {
  return (
    <Provider store={store}>
      <Router />
    </Provider>
  );
};

export default App;
