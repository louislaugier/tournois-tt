import { configureStore } from '@reduxjs/toolkit';
import { keplerGlReducer, enhanceReduxMiddleware } from '@kepler.gl/reducers';
import { MAPBOX_TOKEN } from '../map/constants';

const customizedKeplerGlReducer = keplerGlReducer.initialState({
  uiState: {
    currentModal: null,
    activeSidePanel: 'filter',
    readOnly: true,
    lastActiveSidePanel: 'filter',
  },
  mapStyle: {
    topLayerGroups: {},
    visibleLayerGroups: {
      label: true,
      road: true,
      building: true,
      water: true,
      land: true,
      border: true
    }
  }
});

export const store = configureStore({
  reducer: {
    keplerGl: customizedKeplerGlReducer
  },
  middleware: (getDefaultMiddleware) => enhanceReduxMiddleware(),
  devTools: process.env.NODE_ENV !== 'production'
}); 