import { KeplerGl } from "@kepler.gl/components"
import { MAPBOX_TOKEN } from "../lib/map/constants"
import fr from "../locales/fr"

export default () => {
    return (
        <KeplerGl
            id="paris"
            width={window.innerWidth}
            height={window.innerHeight}
            mapboxApiAccessToken={MAPBOX_TOKEN}
            // override default locale instead of creating a 2nd one
            localeMessages={{ en: fr }}
        />
    )
}