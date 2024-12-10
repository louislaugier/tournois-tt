import React from 'react';
import KeplerGl from '@kepler.gl/components';
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
      width={width}
      height={height}
      mapboxApiAccessToken={process.env.REACT_APP_MAPBOX_TOKEN}
      localeMessages={{ en: fr }} // override EN with FR
    />
  );
}; 