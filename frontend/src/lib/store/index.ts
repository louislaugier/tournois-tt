import { configureStore } from '@reduxjs/toolkit';
import { keplerGlReducer, enhanceReduxMiddleware } from '@kepler.gl/reducers';

const customizedKeplerGlReducer = keplerGlReducer.initialState({
  uiState: {
    currentModal: null,
    activeSidePanel: 'filter'
  }
});

const reducers = {
  keplerGl: customizedKeplerGlReducer,
};

const middlewares = enhanceReduxMiddleware([]);

export const store = configureStore({
  reducer: reducers,
  middleware: (getDefault) => [...middlewares]
}); 