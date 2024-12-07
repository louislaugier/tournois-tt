export interface NominatimResponse {
  place_id: number;
  licence: string;
  osm_type: string;
  osm_id: number;
  boundingbox: string[];
  lat: string;
  lon: string;
  display_name: string;
  class: string;
  type: string;
  importance: number;
  addresstype: string;
  name: string;
  place_rank: number;
}

export interface GeocodingResult {
  latitude: number;
  longitude: number;
  displayName: string;
  boundingBox: {
    minLat: number;
    maxLat: number;
    minLon: number;
    maxLon: number;
  };
}

export class GeocodingError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'GeocodingError';
  }
}

export interface GeocodingOptions {
  countryCode?: string;
  limit?: number;
}

/**
 * Geocodes an address using OpenStreetMap's Nominatim service
 * @param address The address to geocode
 * @param options Optional parameters for geocoding
 * @returns Promise resolving to GeocodingResult or null if not found
 * @throws GeocodingError if the request fails
 */
export async function geocodeAddress(
  address: string,
  options: GeocodingOptions = {}
): Promise<GeocodingResult | null> {
  const {
    countryCode = 'fr',
    limit = 1
  } = options;

  try {
    const response = await fetch(
      `https://nominatim.openstreetmap.org/search?` +
      `format=json` +
      `&q=${encodeURIComponent(address)}` +
      `&limit=${limit}` +
      (countryCode ? `&countrycodes=${countryCode}` : '')
    );

    if (!response.ok) {
      throw new GeocodingError(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json() as NominatimResponse[];

    if (!data || data.length === 0) {
      return null;
    }

    const result = data[0];
    return {
      latitude: parseFloat(result.lat),
      longitude: parseFloat(result.lon),
      displayName: result.display_name,
      boundingBox: {
        minLat: parseFloat(result.boundingbox[0]),
        maxLat: parseFloat(result.boundingbox[1]),
        minLon: parseFloat(result.boundingbox[2]),
        maxLon: parseFloat(result.boundingbox[3])
      }
    };
  } catch (error) {
    if (error instanceof GeocodingError) {
      throw error;
    }
    throw new GeocodingError(
      `Failed to geocode address: ${error instanceof Error ? error.message : 'Unknown error'}`
    );
  }
} 