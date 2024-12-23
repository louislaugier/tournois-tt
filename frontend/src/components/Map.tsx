import React, { useEffect } from 'react';
import KeplerGl from '@kepler.gl/components';
import fr from '../locales/fr';
import { MAPBOX_TOKEN } from '../lib/map/constants';
import { useDispatch } from 'react-redux';
import { toggleSidePanel } from '@kepler.gl/actions';

interface MapProps {
  id: string;
  width: number;
  height: number;
}

export const Map: React.FC<MapProps> = ({ id, width, height }) => {
  const dispatch = useDispatch();

  // Add event listener to handle sidebar state
  useEffect(() => {
    const handleSidebarToggle = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      if (target.closest('.side-panel__close') || target.closest('.side-panel__tab')) {
        // Small delay to let the toggle action complete
        setTimeout(() => {
          dispatch(toggleSidePanel('filter'));
        }, 0);
      }
    };

    document.addEventListener('click', handleSidebarToggle);
    return () => document.removeEventListener('click', handleSidebarToggle);
  }, [dispatch]);

  return (
    <KeplerGl
      id={id}
      width={width}
      height={height}
      mapboxApiAccessToken={MAPBOX_TOKEN}
      localeMessages={{ en: fr }}
    />
  );
}; 