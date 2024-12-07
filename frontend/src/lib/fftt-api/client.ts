import { FFTTQueryParams, Tournament, FFTTResponse } from './types';

const FFTT_API_BASE_URL = 'https://apiv2.fftt.com/api';

export class FFTTError extends Error {
    constructor(message: string) {
        super(message);
        this.name = 'FFTTError';
    }
}

export async function fetchTournaments(params: FFTTQueryParams = {}): Promise<FFTTResponse> {
    const queryParams = new URLSearchParams();
    
    // Add all params to query string with proper encoding
    Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
            const encodedKey = key.replace('[', '%5B').replace(']', '%5D');
            queryParams.append(encodedKey, encodeURIComponent(value.toString()));
        }
    });

    const url = `${FFTT_API_BASE_URL}/tournament_requests?${queryParams.toString()}`;

    try {
        const response = await fetch(url, {
            headers: {
                'Accept': 'application/json',
                'Referer': 'https://monclub.fftt.com/'
            }
        });

        if (!response.ok) {
            throw new FFTTError(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        return data as FFTTResponse;
    } catch (error) {
        if (error instanceof FFTTError) {
            throw error;
        }
        throw new FFTTError(
            `Failed to fetch tournaments: ${error instanceof Error ? error.message : 'Unknown error'}`
        );
    }
}

// Helper function to create date string in the format expected by the API
export function formatDateParam(date: Date): string {
    return date.toISOString().split('.')[0];
}

// New function to handle date parameters
function getDefaultDateParams(): { startDate: Date, endDate: Date } {
    const now = new Date();
    const endOfYear = new Date(now.getFullYear(), 11, 31, 23, 59, 59); // December 31st of current year
    return {
        startDate: now,
        endDate: endOfYear
    };
}

// Query builder for more fluent API usage
export class TournamentQueryBuilder {
    private params: FFTTQueryParams = {};

    page(page: number): this {
        this.params.page = page;
        return this;
    }

    itemsPerPage(count: number): this {
        this.params.itemsPerPage = count;
        return this;
    }

    orderByStartDate(order: 'asc' | 'desc'): this {
        this.params['order[startDate]'] = order;
        return this;
    }

    withStatus(status: number): this {
        this.params.status = status;
        return this;
    }

    withType(type: string): this {
        this.params.type = type;
        return this;
    }

    startDateRange(after?: Date, before?: Date): this {
        if (after) {
            this.params['startDate[after]'] = formatDateParam(after);
        }
        if (before) {
            this.params['startDate[before]'] = formatDateParam(before);
        }
        return this;
    }

    endDateRange(after?: Date, before?: Date): this {
        if (after) {
            this.params['endDate[after]'] = formatDateParam(after);
        }
        if (before) {
            this.params['endDate[before]'] = formatDateParam(before);
        }
        return this;
    }

    inPostalCode(postalCode: string): this {
        this.params['address.postalCode'] = postalCode;
        return this;
    }

    inLocality(locality: string): this {
        this.params['address.addressLocality'] = locality;
        return this;
    }

    async execute(): Promise<FFTTResponse> {
        return fetchTournaments(this.params);
    }

    async executeAll(): Promise<any[]> {
        return fetchAllTournaments(this.params);
    }
}

// Add this new interface to handle pagination metadata
interface FFTTPaginationData {
    totalItems: number;
    itemsPerPage: number;
    totalPages: number;
    currentPage: number;
}

// New function to fetch all tournaments with pagination
export async function fetchAllTournaments(params: FFTTQueryParams = {}): Promise<Tournament[]> {
    const { startDate, endDate } = getDefaultDateParams();

    const defaultParams: FFTTQueryParams = {
        'startDate[after]': formatDateParam(startDate),
        'endDate[before]': formatDateParam(endDate),
        itemsPerPage: 100,
        page: 1,
        ...params
    };

    console.log('Fetching tournaments with params:', defaultParams);

    let allTournaments: Tournament[] = [];
    let currentPage = 1;
    let totalItems = 0;

    do {
        const response = await fetchTournaments({
            ...defaultParams,
            page: currentPage
        });

        if (!response['hydra:member']) {
            console.error('Invalid API response format:', response);
            break;
        }

        const tournaments = response['hydra:member'];
        if (!tournaments.length) break;

        allTournaments = [...allTournaments, ...tournaments];

        // Get total items from first response
        if (currentPage === 1) {
            totalItems = response['hydra:totalItems'] || 0;
            console.log('Total tournaments available:', totalItems);
        }

        // Check if we've fetched all items
        if (allTournaments.length >= totalItems) break;

        currentPage++;

        // Optional: Add a small delay to avoid rate limiting
        await new Promise(resolve => setTimeout(resolve, 100));
    } while (true);

    console.log(`Successfully fetched ${allTournaments.length} tournaments`);
    return allTournaments;
} 