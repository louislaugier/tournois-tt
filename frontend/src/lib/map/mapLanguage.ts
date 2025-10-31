// Utilities to attach French language transformation to a Mapbox GL map instance

const FRENCH_EXPRESSION: any = [
  'coalesce',
  ['get', 'name:fr'],
  ['get', 'name_fr'],
  ['get', 'name:french'],
  ['get', 'name']
];

const IMPORTANT_LAYER_REGEX = /country|state|region|place|settlement|locality|city|town|label/i;

const applyFrenchToLayer = (map: any, layer: any) => {
  if (layer.type !== 'symbol') {
    return false;
  }

  const layerId = layer.id;
  const layout = layer.layout || {};
  const textField = layout['text-field'];

  if (!textField) {
    return false;
  }

  try {
    map.setLayoutProperty(layerId, 'text-field', FRENCH_EXPRESSION);
    return IMPORTANT_LAYER_REGEX.test(layerId);
  } catch (error) {
    return false;
  }
};

const applyFrenchToMap = (map: any) => {
  const style = map.getStyle?.();

  if (!style || !style.layers) {
    return;
  }

  style.layers.forEach((layer: any) => {
    applyFrenchToLayer(map, layer);
  });

  if (typeof map.triggerRepaint === 'function') {
    map.triggerRepaint();
  }
};

export const attachFrenchLanguageToMap = (map: any) => {
  if (!map || typeof map.on !== 'function') {
    return;
  }

  if ((map as any).__tournoisFrenchLanguageApplied) {
    // Already attached
    return;
  }

  (map as any).__tournoisFrenchLanguageApplied = true;

  const apply = () => applyFrenchToMap(map);

  if (map.isStyleLoaded?.()) {
    apply();
  } else {
    map.once('style.load', apply);
  }

  map.on('styledata', () => {
    setTimeout(apply, 100);
  });
};

export const detachFrenchLanguageFromMap = (map: any) => {
  if (!map || !(map as any).__tournoisFrenchLanguageApplied) {
    return;
  }

  delete (map as any).__tournoisFrenchLanguageApplied;
};
