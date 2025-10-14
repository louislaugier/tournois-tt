import { KeplerGl } from "@kepler.gl/components"
import { MAPBOX_TOKEN } from "../lib/map/constants"
import fr from "../locales/fr"
import { LOCALE_CODES } from '@kepler.gl/localization';

export default () => {
    return (
        <KeplerGl
            id="paris"
            width={window.innerWidth}
            height={window.innerHeight}
            mapboxApiAccessToken={MAPBOX_TOKEN}
            // Force French-like strings using our override
            locale={LOCALE_CODES.en}
            localeMessages={{ en: fr }}
        />
    )
}