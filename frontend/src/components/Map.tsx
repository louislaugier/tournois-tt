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

  useEffect(() => {
    let attempts = 0;
    const maxAttempts = 50;

    const setupToggle = () => {
      const sidePanel = document.querySelector('.side-panel');
      const originalToggle = sidePanel?.querySelector('.side-bar__close') as HTMLElement | null;
      
      if (!originalToggle || !sidePanel) {
        if (attempts < maxAttempts) {
          attempts++;
          setTimeout(setupToggle, 100);
        }
        return;
      }

      // Create a new toggle button
      const newToggle = originalToggle.cloneNode(true) as HTMLElement;
      
      // Style the original toggle (hidden but present)
      originalToggle.style.visibility = 'hidden';
      originalToggle.style.position = 'absolute';
      originalToggle.style.pointerEvents = 'none';
      
      // Style the new toggle
      newToggle.style.visibility = 'visible';
      newToggle.style.position = 'absolute';
      newToggle.style.pointerEvents = 'auto';
      
      // Add our custom handler
      const handleToggle = (e: Event) => {
        e.preventDefault();
        e.stopPropagation();

        const container = document.querySelector('.side-panel--container') as HTMLElement;
        const inner = document.querySelector('.side-bar__inner') as HTMLElement;
        
        if (container && inner) {
          if (container.style.width === '0px') {
            container.style.width = '300px';
            inner.style.width = '300px';
            inner.style.opacity = '1';
          } else {
            container.style.width = '0px';
            inner.style.width = '0px';
            inner.style.opacity = '0';
          }
        }
      };

      newToggle.addEventListener('click', handleToggle);

      // Insert the new toggle before the original
      originalToggle.parentElement?.insertBefore(newToggle, originalToggle);

      return () => {
        newToggle.removeEventListener('click', handleToggle);
        newToggle.remove();
        originalToggle.style.visibility = '';
        originalToggle.style.position = '';
        originalToggle.style.pointerEvents = '';
      };
    };

    const cleanup = setupToggle();
    return () => cleanup && cleanup();
  }, []);

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