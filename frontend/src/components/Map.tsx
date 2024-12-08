import React from 'react';
import KeplerGl from '@kepler.gl/components';
import { MAPBOX_TOKEN } from '../lib/map/constants';
import fr from '../locales/fr';

interface MapProps {
  id: string;
  width: number;
  height: number;
}

export const Map: React.FC<MapProps> = ({ id, width, height }) => {
  return (
    <KeplerGl
      id={id}
      mapboxApiAccessToken={MAPBOX_TOKEN}
      width={width}
      height={height}
      localeMessages={{
        en: fr // override en with fr
      }}
    />
  );
}; 