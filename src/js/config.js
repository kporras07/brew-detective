// API Configuration
const API_CONFIG = {
    // Update this URL after deploying to Cloud Run
    BASE_URL: 'https://brew-detective-backend-YOUR_HASH-uc.a.run.app/api/v1',
    
    // For local development
    // BASE_URL: 'http://localhost:8080/api/v1',
    
    ENDPOINTS: {
        CASES: '/cases',
        SUBMISSIONS: '/submissions',
        LEADERBOARD: '/leaderboard',
        USERS: '/users',
        ORDERS: '/orders'
    }
};

// API helper functions
const API = {
    async request(endpoint, options = {}) {
        const url = `${API_CONFIG.BASE_URL}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };
        
        try {
            const response = await fetch(url, config);
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    },

    async get(endpoint) {
        return this.request(endpoint);
    },

    async post(endpoint, data) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    async put(endpoint, data) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    },

    async delete(endpoint) {
        return this.request(endpoint, {
            method: 'DELETE'
        });
    }
};