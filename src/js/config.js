// API Configuration
const API_CONFIG = {
    // Production backend URL on Cloud Run
    BASE_URL: 'https://brew-detective-backend-1087966598090.us-central1.run.app',
    
    // For local development
    // BASE_URL: 'http://localhost:8888',
    
    ENDPOINTS: {
        // API endpoints
        CASES: '/api/v1/cases',
        ACTIVE_CASE: '/api/v1/cases/active',
        ACTIVE_CASE_PUBLIC: '/api/v1/cases/active/public', // Secure endpoint without answers
        SUBMISSIONS: '/api/v1/submissions',
        LEADERBOARD: '/api/v1/leaderboard',
        LEADERBOARD_CURRENT: '/api/v1/leaderboard/current',
        USERS: '/api/v1/users',
        ORDERS: '/api/v1/orders',
        PROFILE: '/api/v1/profile',
        CATALOG: '/api/v1/catalog',
        
        // Admin endpoints
        ADMIN_CATALOG: '/api/v1/admin/catalog',
        ADMIN_CASES: '/api/v1/admin/cases',
        ADMIN_ORDERS: '/api/v1/admin/orders',
        ADMIN_USERS: '/api/v1/admin/users',
        
        // Auth endpoints
        AUTH_GOOGLE: '/auth/google',
        AUTH_CALLBACK: '/auth/google/callback',
        AUTH_LOGOUT: '/auth/logout'
    }
};

// Authentication helper (global)
window.Auth = {
    getToken() {
        return localStorage.getItem('auth_token');
    },

    setToken(token) {
        localStorage.setItem('auth_token', token);
    },

    removeToken() {
        localStorage.removeItem('auth_token');
        localStorage.removeItem('user_data');
    },

    getUser() {
        const userData = localStorage.getItem('user_data');
        return userData ? JSON.parse(userData) : null;
    },

    setUser(user) {
        localStorage.setItem('user_data', JSON.stringify(user));
    },

    isAuthenticated() {
        const token = this.getToken();
        if (!token) return false;
        
        try {
            // Basic JWT validation (check if token is expired)
            const payload = JSON.parse(atob(token.split('.')[1]));
            return payload.exp * 1000 > Date.now();
        } catch {
            return false;
        }
    },

    async login() {
        try {
            const response = await fetch(`${API_CONFIG.BASE_URL}${API_CONFIG.ENDPOINTS.AUTH_GOOGLE}`);
            const data = await response.json();
            window.location.href = data.auth_url;
        } catch (error) {
            console.error('Login failed:', error);
            throw error;
        }
    },

    async logout() {
        try {
            if (this.isAuthenticated()) {
                await API.post(API_CONFIG.ENDPOINTS.AUTH_LOGOUT);
            }
        } catch (error) {
            console.error('Logout request failed:', error);
        } finally {
            this.removeToken();
            window.location.reload();
        }
    }
};

// API helper functions
const API = {
    async request(endpoint, options = {}) {
        const url = `${API_CONFIG.BASE_URL}${endpoint}`;
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers
        };

        // Add auth token if available
        const token = Auth.getToken();
        if (token) {
            headers.Authorization = `Bearer ${token}`;
        }

        const config = {
            headers,
            ...options
        };
        
        try {
            const response = await fetch(url, config);
            
            if (response.status === 401) {
                // Token expired or invalid
                Auth.removeToken();
                throw new Error('Authentication required');
            }
            
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
