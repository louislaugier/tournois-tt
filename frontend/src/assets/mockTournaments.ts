import { Tournament } from "../lib/api/types";

// mock = manually added tournaments

export const mockTournaments: Array<Tournament> = [
   
];

export const mockPastCurrentTournaments: Array<Tournament> = [
    
]

export const mockPastTournaments: Array<Tournament> = [
    {
        affiche: 'https://cdn.paris.fr/paris/2025/02/10/original-8d98985523369228222740f1675f4a0e.png',
        '@id': '/tournaments/mock-att-xv-2025',
        '@type': 'Tournament',
        id: 999996, // Unique mock ID
        identifier: 'MOCK-ATT-XV-2025',
        name: "Les Olymping's du 15e",
        type: 'Départemental',
        club: {
            '@id': '/clubs/attxv',
            '@type': 'Club',
            id: 68, // Unique mock ID
            name: 'ASSOC. TENNIS DE TABLE PARIS XVe',
            identifier: '08751260'
        },
        startDate: new Date('2025-03-22').toISOString(),
        endDate: new Date('2025-03-23').toISOString(),
        address: {
            '@id': '/addresses/mairie-xv',
            '@type': 'PostalAddress',
            id: 999997, // Unique mock ID
            postalCode: '75015',
            streetAddress: '31 rue Péclet',
            disambiguatingDescription: 'Mairie du XVème - Salle des fêtes',
            addressCountry: 'FR',
            addressRegion: 'Ile-de-France',
            addressLocality: 'Paris',
            areaServed: null,
            latitude: 48.8411737,
            longitude: 2.2991291,
            name: 'Mairie du XVème - Salle des fêtes',
            identifier: null,
            openingHours: null,
            main: false,
            approximate: false
        },
        contacts: [], // Empty contacts array
        rules: {
            url: ""
        },
        endowment: 70000, // Total endowment in cents
        organization: undefined, // Optional field
        responses: [], // Optional field
        engagmentSheet: undefined, // Optional field
        decision: undefined, // Optional field
        page: null, // Optional field
        '@permissions': {
            canUpdate: true,
            canDelete: false
        }
    },
];