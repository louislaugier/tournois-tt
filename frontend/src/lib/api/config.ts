export const API_BASE_URL = '/api/v1'; // Using relative path for proxy 

// Get API key from environment variables
const API_KEY = process.env.REACT_APP_API_KEY;

// Headers used for all API requests
export const getDefaultHeaders = () => ({
    'Accept': 'application/json',
    'X-API-Key': API_KEY || '',
});