export const DEFAULT_MAP_CONFIG = {
    visState: {
        filters: [],
        interactionConfig: {
            tooltip: {
                enabled: true,
                fieldsToShow: {
                    tournament_data: [
                        { name: 'name', format: null }
                    ]
                },
                compareMode: false,
                compareType: 'absolute'
            },
            brush: {
                enabled: false
            },
            coordinate: {
                enabled: false
            },
            geocoder: {
                enabled: true
            }
        },
        layerBlending: 'normal',
        splitMaps: [],
        animationConfig: {
            currentTime: null,
            speed: 1
        },
        layers: [{
            id: 'point',
            type: 'point',
            config: {
                dataId: 'tournament_data',
                label: 'Tournoi',
                color: [18, 147, 154] as [number, number, number],
                columns: {
                    lat: 'latitude',
                    lng: 'longitude'
                },
                isVisible: true,
                visConfig: {
                    radius: 10,
                    fixedRadius: false,
                    opacity: 0.8,
                    outline: false,
                    filled: true
                }
            }
        }]
    },
    mapState: {
        latitude: 46.777138,
        longitude: 2.804568,
        bearing: 0,
        pitch: 0,
        zoom: 5.6,
        dragRotate: false
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
}; 