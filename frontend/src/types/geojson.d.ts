declare module '*.geojson' {
  const content: {
    type: string;
    geometry: {
      type: string;
      coordinates: any[];
    };
  };
  export default content;
} 