declare module 's2-geometry' {
  export class S2LatLng {
    static fromDegrees(lat: number, lng: number): S2LatLng;
  }

  export class S2CellId {
    static fromLatLng(latlng: S2LatLng): S2CellId;
    toString(): string;
  }
} 