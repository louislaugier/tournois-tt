import { getCurrentSeasonYears } from "../utils/season";

export const getMapLayers = () => {
    const { seasonStartYear, seasonEndYear } = getCurrentSeasonYears()

    return [
        createLayerConfig(
            'current_tournaments',
            [31, 186, 214],
            `Saison en cours - tournoi à venir (${seasonStartYear}-${seasonEndYear})`,
            16,  // text label size
            24   // radius
        ),
        createLayerConfig(
            'past_current_tournaments',
            [0, 0, 139],
            `Saison en cours - tournoi terminé (${seasonStartYear}-${seasonEndYear})`,
            16,  // text label size
            22   // radius
        ),
        createLayerConfig(
            'past_tournaments',
            [155, 89, 182],
            `Saison précédente (${seasonStartYear - 1}-${seasonEndYear - 1})`,
            16,  // text label size
            26   // radius
        )
    ]
}

const createLayerConfig = (
    id: string,
    color: [number, number, number],
    label: string,
    textLabelSize: number = 16,
    radius: number = 20
) => {
    return {
        id,
        type: 'point',
        config: {
            dataId: id,
            label,
            color,
            columns: DEFAULT_COLUMNS,
            isVisible: true,
            visConfig: {
                ...DEFAULT_VIS_CONFIG,
                radius
            },
            textLabel: {
                ...DEFAULT_TEXT_LABEL_CONFIG,
                field: { name: 'count', type: 'string' },
                size: textLabelSize,
            },
        },
        visualChannels: DEFAULT_VISUAL_CHANNELS
    };
};

const DEFAULT_TEXT_LABEL_CONFIG = {
    color: [255, 255, 255] as [number, number, number],
    offset: [0, 0] as [number, number],
    anchor: 'middle',
    alignment: 'center',
    background: true,
    backgroundColor: [0, 0, 0, 0.5] as [number, number, number, number],
    outlineWidth: 0,
    outlineColor: [0, 0, 0, 0.5] as [number, number, number, number]
};

const DEFAULT_VIS_CONFIG = {
    fixedRadius: false,
    opacity: 0.8,
    outline: false,
    filled: true,
    radiusRange: [20, 30]
};

const DEFAULT_VISUAL_CHANNELS = {
    colorField: null,
    colorScale: 'quantile',
    sizeScale: 'linear',
    strokeColorField: null,
    strokeColorScale: 'quantile'
};

const DEFAULT_COLUMNS = {
    lat: 'latitude',
    lng: 'longitude'
} as { [key: string]: string };