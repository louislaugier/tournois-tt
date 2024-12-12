export const DEFAULT_MAP_CONFIG = {
    visState: {
        filters: [],
        interactionConfig: {
            tooltip: {
                enabled: true,
                fieldsToShow: {
                    tournament_data: [
                        // { name: 'latitude', format: null },
                        // { name: 'longitude', format: null },
                        {
                            name: 'Localisation',
                            format: null,
                        },
                        { name: 'Nom du tournoi', format: null },
                        { name: 'Type', format: null },
                        { name: 'Club', format: null },
                        { name: 'Dotation totale', format: null },
                        { name: 'Dates de début / fin', format: null },
                        {
                            name: 'Adresse',
                            format: 'Voir sur Google Maps',
                            type: 'link',
                        },
                        {
                            name: 'Règlement',
                            format: 'Voir le règlement',
                            type: 'link',
                        }
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
            id: 'tournoi',
            type: 'point',
            config: {
                dataId: 'tournament_data',
                label: 'Tournoi',
                color: [51, 153, 255] as [number, number, number],
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