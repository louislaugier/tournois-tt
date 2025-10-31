import React, { useCallback } from "react";
import { KeplerGl } from "@kepler.gl/components";
import { MAPBOX_TOKEN } from "../lib/map/constants";
import fr from "../locales/fr";
import { LOCALE_CODES } from '@kepler.gl/localization';
import { attachFrenchLanguageToMap } from "../lib/map/mapLanguage";

const MapComponent: React.FC = () => {
    const handleGetMapboxRef = useCallback((mapbox?: any) => {
        if (!mapbox) {
            return;
        }

        const map = typeof mapbox.getMap === 'function' ? mapbox.getMap() : mapbox;
        if (map) {
            attachFrenchLanguageToMap(map);
        }
    }, []);

    return (
        <KeplerGl
            id="paris"
            width={window.innerWidth}
            height={window.innerHeight}
            mapboxApiAccessToken={MAPBOX_TOKEN}
            // Force French-like strings using our override
            locale={LOCALE_CODES.en}
            localeMessages={{ en: fr }}
            getMapboxRef={handleGetMapboxRef}
        />
    );
};

export default MapComponent;