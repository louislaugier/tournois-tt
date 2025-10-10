import React from 'react';
import { Provider } from 'react-redux';
import '@kepler.gl/styles';
import { HeroUIProvider } from '@heroui/react';
import { store } from './lib/store';
import Router from './components/Router';

const App: React.FC = () => {
  return (
    <Provider store={store}>
      <HeroUIProvider>
        <Router />
      </HeroUIProvider>
    </Provider>
  );
};

export default App;
