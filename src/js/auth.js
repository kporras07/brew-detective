// OAuth callback handler
async function handleOAuthCallback() {
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');
    const state = urlParams.get('state');
    const error = urlParams.get('error');

    if (error) {
        console.error('OAuth error:', error);
        alert('Error al iniciar sesión. Por favor intenta nuevamente.');
        window.location.href = '/';
        return;
    }

    if (code && state) {
        // Send the code to our backend
        fetch(`${API_CONFIG.BASE_URL}${API_CONFIG.ENDPOINTS.AUTH_CALLBACK}?code=${code}&state=${state}`)
            .then(response => response.json())
            .then(async data => {
                if (data.token && data.user) {
                    Auth.setToken(data.token);
                    Auth.setUser(data.user);
                    await updateAuthUI();
                    showNotification('¡Sesión iniciada exitosamente!', 'success');
                    // Redirect to home
                    window.location.href = '/';
                } else {
                    throw new Error('Invalid response from server');
                }
            })
            .catch(error => {
                console.error('Callback error:', error);
                alert('Error al procesar la autenticación. Por favor intenta nuevamente.');
                window.location.href = '/';
            });
    }
}

// Update authentication UI elements
async function updateAuthUI() {
    const isAuthenticated = Auth.isAuthenticated();
    const user = Auth.getUser();

    console.log('Updating auth UI:', { isAuthenticated, user }); // Debug

    // Desktop navigation
    const loginBtn = document.getElementById('loginBtn');
    const userDropdown = document.getElementById('userDropdown');
    const userName = document.getElementById('userName');
    const submitMenuItem = document.getElementById('submitMenuItem');
    
    // Mobile navigation
    const loginBtnMobile = document.getElementById('loginBtnMobile');
    const userDropdownMobile = document.getElementById('userDropdownMobile');
    const submitMenuItemMobile = document.getElementById('submitMenuItemMobile');

    console.log('Elements found:', { loginBtn, userDropdown, userName }); // Debug

    if (isAuthenticated && user) {
        console.log('User is authenticated, showing user dropdown'); // Debug
        
        // Desktop: Hide login button, show user dropdown
        if (loginBtn) {
            loginBtn.style.display = 'none';
        }
        if (userDropdown) {
            userDropdown.style.display = 'block';
            console.log('User dropdown shown'); // Debug
        }
        if (userName) {
            userName.textContent = user.name || user.email;
            console.log('User name set to:', user.name || user.email); // Debug
        }

        // Mobile: Hide login button, show user dropdown
        if (loginBtnMobile) {
            loginBtnMobile.style.display = 'none';
        }
        if (userDropdownMobile) {
            userDropdownMobile.style.display = 'block';
        }

        // Show submit menu items for authenticated users
        if (submitMenuItem) {
            submitMenuItem.style.display = 'block';
        }
        if (submitMenuItemMobile) {
            submitMenuItemMobile.style.display = 'block';
        }

        // Update profile page with user data
        await updateProfilePage(user);
    } else {
        console.log('User not authenticated, showing login button'); // Debug
        
        // Desktop: Show login button, hide user dropdown
        if (loginBtn) {
            loginBtn.style.display = 'inline';
        }
        if (userDropdown) {
            userDropdown.style.display = 'none';
        }

        // Mobile: Show login button, hide user dropdown
        if (loginBtnMobile) {
            loginBtnMobile.style.display = 'inline';
        }
        if (userDropdownMobile) {
            userDropdownMobile.style.display = 'none';
        }

        // Hide submit menu items for unauthenticated users
        if (submitMenuItem) {
            submitMenuItem.style.display = 'none';
        }
        if (submitMenuItemMobile) {
            submitMenuItemMobile.style.display = 'none';
        }
    }
}

// Update profile page with user data
async function updateProfilePage(user) {
    // Update profile form fields
    const profileName = document.getElementById('profileName');
    const profileEmail = document.getElementById('profileEmail');

    if (profileName) {
        profileName.value = user.name || '';
    }
    if (profileEmail) {
        profileEmail.value = user.email || '';
        profileEmail.disabled = true; // Email from Google shouldn't be editable
    }

    // Update statistics with real data from API
    await updateUserStats(user);
}

// Update user statistics
async function updateUserStats(user) {
    // Show loading state for stats
    const pointsDisplay = document.querySelector('[data-stat="points"]');
    const rankDisplay = document.querySelector('[data-stat="rank"]');
    const casesDisplay = document.querySelector('[data-stat="cases"]');
    const accuracyDisplay = document.querySelector('[data-stat="accuracy"]');

    // Set loading state
    if (pointsDisplay) pointsDisplay.textContent = '...';
    if (rankDisplay) rankDisplay.textContent = '...';
    if (casesDisplay) casesDisplay.textContent = '...';
    if (accuracyDisplay) accuracyDisplay.textContent = '...';
    
    try {
        // Fetch real user rank from API
        const leaderboardResponse = await API.get(API_CONFIG.ENDPOINTS.LEADERBOARD);
        const leaderboard = leaderboardResponse.leaderboard || [];
        
        // Find user's actual rank in leaderboard
        const userRank = leaderboard.findIndex(entry => entry.user_id === user.id) + 1;
        
        const stats = {
            points: user.score || 0,
            rank: userRank > 0 ? userRank : '---',
            casesSolved: user.cases_solved || 0,
            accuracy: Math.round((user.accuracy || 0) * 100)
        };

        // Update UI elements with actual data
        if (pointsDisplay) pointsDisplay.textContent = stats.points.toLocaleString();
        if (rankDisplay) rankDisplay.textContent = stats.rank !== '---' ? `# ${stats.rank}` : 'Sin ranking';
        if (casesDisplay) casesDisplay.textContent = stats.casesSolved;
        if (accuracyDisplay) accuracyDisplay.textContent = `${stats.accuracy}%`;
    } catch (error) {
        console.error('Failed to load user stats:', error);
        // Show error state instead of fake data
        if (pointsDisplay) pointsDisplay.textContent = 'Error';
        if (rankDisplay) rankDisplay.textContent = 'Error';
        if (casesDisplay) casesDisplay.textContent = 'Error';
        if (accuracyDisplay) accuracyDisplay.textContent = 'Error';
    }
}

// Show notification
function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `notification notification--${type}`;
    notification.textContent = message;
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 15px 20px;
        border-radius: 8px;
        color: white;
        font-weight: bold;
        z-index: 1000;
        opacity: 0;
        transition: opacity 0.3s ease;
    `;

    switch (type) {
        case 'success':
            notification.style.backgroundColor = '#2ecc71';
            break;
        case 'error':
            notification.style.backgroundColor = '#e74c3c';
            break;
        default:
            notification.style.backgroundColor = '#3498db';
    }

    document.body.appendChild(notification);
    
    // Fade in
    setTimeout(() => {
        notification.style.opacity = '1';
    }, 100);

    // Fade out and remove
    setTimeout(() => {
        notification.style.opacity = '0';
        setTimeout(() => {
            document.body.removeChild(notification);
        }, 300);
    }, 3000);
}

// Check if we're on the OAuth callback or have a token in the URL
(async () => {
    if (window.location.pathname === '/auth/callback' || window.location.search.includes('code=')) {
        await handleOAuthCallback();
    } else if (window.location.hash.includes('token=')) {
        await handleTokenFromURL();
    }
})();

// Handle token received from backend redirect
async function handleTokenFromURL() {
    const hash = window.location.hash.substring(1);
    const params = new URLSearchParams(hash);
    const token = params.get('token');

    if (token) {
        Auth.setToken(token);
        
        // Get user data from the token
        try {
            const payload = JSON.parse(atob(token.split('.')[1]));
            const userData = {
                id: payload.user_id,
                email: payload.email,
                name: payload.name
            };
            Auth.setUser(userData);
        } catch (error) {
            console.error('Failed to parse token:', error);
        }

        await updateAuthUI();
        showNotification('¡Sesión iniciada exitosamente!', 'success');
        
        // Clean up the URL
        window.history.replaceState({}, document.title, window.location.pathname);
        
        // Show the profile page
        showPage('profile');
    }
}

// Initialize auth UI when page loads
document.addEventListener('DOMContentLoaded', async function() {
    await updateAuthUI();
});

// Manual test function for debugging
window.testAuthUI = function() {
    console.log('=== AUTH UI TEST ===');
    console.log('Auth.isAuthenticated():', Auth.isAuthenticated());
    console.log('Auth.getUser():', Auth.getUser());
    console.log('Auth.getToken():', Auth.getToken() ? 'Present' : 'Missing');
    
    const elements = {
        loginBtn: document.getElementById('loginBtn'),
        userDropdown: document.getElementById('userDropdown'),
        userName: document.getElementById('userName'),
        loginBtnMobile: document.getElementById('loginBtnMobile'),
        userDropdownMobile: document.getElementById('userDropdownMobile')
    };
    
    console.log('DOM Elements:', elements);
    
    Object.keys(elements).forEach(key => {
        const el = elements[key];
        if (el) {
            console.log(`${key} styles:`, {
                display: el.style.display,
                visibility: el.style.visibility,
                opacity: el.style.opacity
            });
        }
    });
    
    console.log('=== END TEST ===');
};