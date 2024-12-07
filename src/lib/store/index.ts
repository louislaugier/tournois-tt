import { configureStore } from '@reduxjs/toolkit';
import { keplerGlReducer, enhanceReduxMiddleware } from '@kepler.gl/reducers';

// Create a customized reducer
const customizedKeplerGlReducer = keplerGlReducer.initialState({
  uiState: {
    // Add any UI state customization here
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