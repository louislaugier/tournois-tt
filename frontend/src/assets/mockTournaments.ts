import { Tournament } from "../lib/api/types";

// mock = manually added tournaments

export const mockTournaments: Array<Tournament> = [
    // {
    //     affiche: 'https://cdn.pepsup.com/resources/images/ARTICLES/000/184/552/1845523/IMAGE/1845523.jpg?1732633837000',
    //     '@id': '/tournaments/mock-kb-2025',
    //     '@type': 'Tournament',
    //     id: 999999, // Unique mock ID
    //     identifier: 'MOCK-KB-2025',
    //     name: 'TOURNOI NATIONAL DU KREMLIN-BICETRE',
    //     type: 'National B',
    //     club: {
    //         '@id': '/clubs/uskb',
    //         '@type': 'Club',
    //         id: 69, // Unique mock ID
    //         name: 'KREMLIN-BICETRE T.T.',
    //         identifier: '08940975'
    //     },
    //     startDate: new Date('2025-06-14').toISOString(),
    //     endDate: new Date('2025-06-15').toISOString(),
    //     address: {
    //         '@id': '/addresses/kb-gymnase',
    //         '@type': 'PostalAddress',
    //         id: 999999, // Unique mock ID
    //         postalCode: '94270',
    //         streetAddress: '12 bd Chastenet de Géry',
    //         disambiguatingDescription: 'COSEC Elisabeth et Vincent Purkart',
    //         addressCountry: 'FR',
    //         addressRegion: 'Ile-de-France',
    //         addressLocality: 'Le Kremlin-Bicêtre',
    //         areaServed: null,
    //         latitude: 48.807173,
    //         longitude: 2.35724,
    //         name: 'COSEC Elisabeth et Vincent Purkart',
    //         identifier: null,
    //         openingHours: null,
    //         main: false,
    //         approximate: false
    //     },
    //     contacts: [], // Empty contacts array
    //     rules: {
    //         url: ""
    //     },
    //     endowment: 0, // Total endowment in cents
    //     organization: undefined, // Optional field
    //     responses: [], // Optional field
    //     engagmentSheet: undefined, // Optional field
    //     decision: undefined, // Optional field
    //     page: null, // Optional field
    //     '@permissions': {
    //         canUpdate: true,
    //         canDelete: false
    //     }
    // },
    // {
    //     affiche: 'https://www.esvitrytt.fr/kcfinder/upload/files/ES%20VITRY%20TT%20TOURNOI%20NATIONAL%20B.pdf',
    //     '@id': '/tournaments/mock-vitry-2025',
    //     '@type': 'Tournament',
    //     id: 999998, // Unique mock ID
    //     identifier: 'MOCK-VITRY-2025',
    //     name: 'ES VITRY TENNIS DE TABLE TOURNOI NATIONAL B',
    //     type: 'National B',
    //     club: {
    //         '@id': '/clubs/esvitry',
    //         '@type': 'Club',
    //         id: 68, // Unique mock ID
    //         name: 'VITRY ES',
    //         identifier: '08940448'
    //     },
    //     startDate: new Date('2025-04-26').toISOString(),
    //     endDate: new Date('2025-04-27').toISOString(),
    //     address: {
    //         '@id': '/addresses/vitry-gymnase',
    //         '@type': 'PostalAddress',
    //         id: 999998, // Unique mock ID
    //         postalCode: '94400',
    //         streetAddress: '4 Avenue du Colonel Fabien',
    //         disambiguatingDescription: 'GYMNASE GOSNAT',
    //         addressCountry: 'FR',
    //         addressRegion: 'Ile-de-France',
    //         addressLocality: 'Vitry-sur-Seine',
    //         areaServed: null,
    //         latitude: 48.7815592,
    //         longitude: 2.3802291,
    //         name: 'GYMNASE GOSNAT',
    //         identifier: null,
    //         openingHours: null,
    //         main: false,
    //         approximate: false
    //     },
    //     contacts: [], // Empty contacts array
    //     rules: {
    //         url: ""
    //     },
    //     endowment: 131000, // Total endowment in cents
    //     organization: undefined, // Optional field
    //     responses: [], // Optional field
    //     engagmentSheet: undefined, // Optional field
    //     decision: undefined, // Optional field
    //     page: null, // Optional field
    //     '@permissions': {
    //         canUpdate: true,
    //         canDelete: false
    //     }
    // },
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
    }
];


export const mockPastTournaments: Array<Tournament> = [
    {
        affiche: '',
        '@id': '/tournaments/mock-att-xv-2024',
        '@type': 'Tournament',
        id: 999995, // Unique mock ID
        identifier: 'MOCK-ATT-XV-2024',
        name: "Les Olymping's du 15e",
        type: 'Départemental',
        club: {
            '@id': '/clubs/attxv',
            '@type': 'Club',
            id: 68, // Unique mock ID
            name: 'ASSOC. TENNIS DE TABLE PARIS XVe',
            identifier: '08751260'
        },
        startDate: new Date('2024-03-24').toISOString(),
        endDate: new Date('2024-03-24').toISOString(),
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
            url: "https://attxv.fr/wp-content/uploads/2024/03/2024_Olympings_Reglement-Tournoi-XV-adulte.pdf"
        },
        endowment: 60000, // Total endowment in cents
        organization: undefined, // Optional field
        responses: [], // Optional field
        engagmentSheet: undefined, // Optional field
        decision: undefined, // Optional field
        page: null, // Optional field
        '@permissions': {
            canUpdate: true,
            canDelete: false
        }
    }
];