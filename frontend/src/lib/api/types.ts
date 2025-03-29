export interface Address {
  '@id': string;
  '@type': string;
  postalCode: string;
  streetAddress: string;
  disambiguatingDescription: string | null;
  addressCountry: string | null;
  addressRegion: string | null;
  addressLocality: string;
  areaServed: string | null;
  latitude: number | null;
  longitude: number | null;
  name: string | null;
  identifier: string | null;
  openingHours: string | null;
  main: boolean;
  id: number;
  approximate?: boolean;
}

export interface Club {
  '@id': string;
  '@type': string;
  name: string;
  identifier: string;
  id: number;
}

export interface Contact {
  '@id': string;
  '@type': string;
  type: string;
  givenName: string;
  familyName: string;
  email: string;
  telephone: string;
  id: number;
}

export interface Table {
  name: string;
  description: string;
  date: string;
  time: string;
  fee: number;
  endowment: number;
}

export interface Organization {
  '@id': string;
  '@type': string;
  name: string;
  identifier: string;
  id: number;
}

export interface Tournament {
  affiche?: string;
  '@id': string;
  '@type': string;
  club: Club;
  identifier: string;
  name: string;
  type: string;
  startDate: string;
  endDate: string;
  address: Address;
  contacts: Contact[];
  rules?: {
    url: string;
  };
  tables?: Table[];
  endowment: number;
  id: number;
  signupUrl?: string;
  organization?: {
    '@id': string;
    '@type': string;
    name: string;
    identifier: string;
    id: number;
  };
  responses?: {
    '@id': string;
    '@type': string;
    accountant: string;
    date: string;
    review: number;
    description: string | null;
    organization: Organization;
    id: number;
  }[];
  engagmentSheet?: {
    '@id': string;
    '@type': string;
    originalFilename: string;
    mimeType: string;
    size: number;
    id: number;
    url: string;
  };
  decision?: any;
  page?: string | null;
  '@permissions'?: {
    canUpdate: boolean;
    canDelete: boolean;
  };
}

export interface FFTTResponse {
  '@context': string;
  '@id': string;
  '@type': string;
  'hydra:member': Tournament[];
  'hydra:totalItems': number;
}

export interface FFTTQueryParams {
  page?: number;
  itemsPerPage?: number;
  'order[startDate]'?: 'asc' | 'desc';
  status?: number;
  type?: string;
  'startDate[after]'?: string;
  'startDate[before]'?: string;
  'endDate[after]'?: string;
  'endDate[before]'?: string;
  'address.postalCode'?: string;
  'address.addressLocality'?: string;
}

export interface Response {
  '@id': string;
  '@type': string;
  accountant: string;
  date: string;
  review: number;
  description: string | null;
  organization: Organization;
  id: number;
} 