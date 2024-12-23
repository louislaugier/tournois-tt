import { configureStore } from '@reduxjs/toolkit';
import { keplerGlReducer, enhanceReduxMiddleware } from '@kepler.gl/reducers';
import { MAPBOX_TOKEN } from '../map/constants';

const customizedKeplerGlReducer = keplerGlReducer.initialState({
  uiState: {
    currentModal: null,
    activeSidePanel: 'filter',
    readOnly: true,
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
  middleware: (getDefaultMiddleware) =>
    enhanceReduxMiddleware(
      // getDefaultMiddleware({
      //   serializableCheck: {
      //     // Ignore these action types
      //     ignoredActions: [
      //       // '@@kepler.gl/REGISTER_ENTRY',
      //       // '@@kepler.gl/ADD_DATA_TO_MAP',
      //       // '@@kepler.gl/MOUSE_MOVE',
      //       // '@@kepler.gl/UPDATE_MAP',
      //       // '@@kepler.gl/LAYER_HOVER',
      //       // '@@kepler.gl/LOAD_MAP_STYLES'
      //     ],
      //     // Ignore these field paths in the state
      //     ignoredPaths: [
      //       // 'keplerGl.paris.visState.layerClasses',
      //       // 'keplerGl.paris.visState.layers',
      //       // 'payload.info.viewport',
      //       // 'payload.evt',
      //       // 'payload.payload'
      //     ],
      //     // Increase the warning threshold for performance
      //     warnAfter: 200
      //   },
      //   immutableCheck: {
      //     // Increase the warning threshold for performance
      //     warnAfter: 200
      //   }
      // })
    ),
  devTools: process.env.NODE_ENV !== 'production'
}); 