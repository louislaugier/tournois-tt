export const API_BASE_URL = '/api/v1'; // Using relative path for proxy 

// Headers used for all API requests
export const getDefaultHeaders = () => ({
    'Content-Type': 'application/json',
    'Accept': 'application/json',
});