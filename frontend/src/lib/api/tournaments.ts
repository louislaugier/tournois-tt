import { formatDateQueryParam } from '../utils/date';
import { API_BASE_URL, getDefaultHeaders } from './config';
import { Tournament } from './types';

export class APIError extends Error {
    constructor(message: string) {
        super(message);
        this.name = 'Internal API error';
    }
}

export interface TournamentQueryParams {
    page?: number;
    itemsPerPage?: number;
    'order[startDate]'?: 'asc' | 'desc';
    'startDate[after]'?: string;
    'startDate[before]'?: string;
    'endDate[after]'?: string;
    'endDate[before]'?: string;
    'address.postalCode'?: string;
    'address.addressLocality'?: string;
    type?: string;
}

export async function fetchTournaments(params: TournamentQueryParams = {}): Promise<Tournament[]> {
    const queryParams = new URLSearchParams();

    Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
            queryParams.append(key, value.toString());
        }
    });

    const url = `${API_BASE_URL}/tournaments?${queryParams.toString()}`;

    try {
        const response = await fetch(url, {
            headers: getDefaultHeaders()
        });

        if (!response.ok) {
            throw new APIError(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        return data;
    } catch (error) {
        if (error instanceof APIError) {
            throw error;
        }
        throw new APIError(
            `Failed to fetch tournaments: ${error instanceof Error ? error.message : 'Unknown error'}`
        );
    }
}

export async function fetchAllTournaments(params: TournamentQueryParams = {}): Promise<Tournament[]> {
    const defaultParams: TournamentQueryParams = {
        itemsPerPage: 100,
        page: 1,
        ...params
    };

    try {
        const tournaments = await fetchTournaments(defaultParams);
        return tournaments;
    } catch (error) {
        console.error('Error fetching tournaments:', error);
        return [];
    }
}

// Query builder for more fluent API usage
export class TournamentQueryBuilder {
    private params: TournamentQueryParams = {};

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

    withType(type: string): this {
        this.params.type = type;
        return this;
    }

    startDateRange(after?: Date, before?: Date): this {
        if (after) {
            this.params['startDate[after]'] = formatDateQueryParam(after);
        }
        if (before) {
            this.params['startDate[before]'] = formatDateQueryParam(before);
        }
        return this;
    }

    endDateRange(after?: Date, before?: Date): this {
        if (after) {
            this.params['endDate[after]'] = formatDateQueryParam(after);
        }
        if (before) {
            this.params['endDate[before]'] = formatDateQueryParam(before);
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

    async execute(): Promise<Tournament[]> {
        const results = await fetchTournaments(this.params);
        return results;
    }

    async executeAndLogAll(): Promise<Tournament[]> {
        const tournaments = await fetchAllTournaments(this.params);
        return tournaments;
    }
} 