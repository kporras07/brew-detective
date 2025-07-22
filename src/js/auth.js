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
    let user = Auth.getUser();

    // If user is authenticated, fetch fresh user data from API to get latest type
    if (isAuthenticated && user) {
        try {
            const freshUserData = await API.get(API_CONFIG.ENDPOINTS.PROFILE);
            user = freshUserData;
            Auth.setUser(user); // Update cached user data
            // Fresh user data fetched
        } catch (error) {
            console.error('Failed to fetch fresh user data:', error);
            // Continue with cached user data
        }
    }

    // Updating authentication UI

    // Desktop navigation
    const loginBtn = document.getElementById('loginBtn');
    const userDropdown = document.getElementById('userDropdown');
    const userName = document.getElementById('userName');
    const submitMenuItem = document.getElementById('submitMenuItem');
    
    // Mobile navigation
    const loginBtnMobile = document.getElementById('loginBtnMobile');
    const userDropdownMobile = document.getElementById('userDropdownMobile');
    const submitMenuItemMobile = document.getElementById('submitMenuItemMobile');

    // DOM elements located

    if (isAuthenticated && user) {
        // User authenticated, updating UI
        
        // Desktop: Hide login button, show user dropdown
        if (loginBtn) {
            loginBtn.style.display = 'none';
        }
        if (userDropdown) {
            userDropdown.classList.remove('hidden');
            // User dropdown displayed
        }
        if (userName) {
            userName.textContent = user.name || user.email;
            // User name updated
        }
        
        // Update user avatar
        const userAvatar = document.getElementById('userAvatar');
        if (userAvatar && user.picture) {
            userAvatar.src = user.picture;
            userAvatar.style.display = 'inline-block';
            // User avatar updated
        }

        // Mobile: Hide login button, show user dropdown
        if (loginBtnMobile) {
            loginBtnMobile.style.display = 'none';
        }
        if (userDropdownMobile) {
            userDropdownMobile.classList.remove('hidden');
        }

        // Show submit menu items for authenticated users
        if (submitMenuItem) {
            submitMenuItem.style.display = 'block';
        }
        if (submitMenuItemMobile) {
            submitMenuItemMobile.style.display = 'block';
        }

        // Show admin menu item for admin users
        const adminMenuItem = document.getElementById('adminMenuItem');
        if (adminMenuItem) {
            // Checking admin access
            if (user.type === 'admin') {
                // Admin menu displayed
                adminMenuItem.style.display = 'block';
            } else {
                // Admin menu hidden
                adminMenuItem.style.display = 'none';
            }
        }

        // Update profile page with user data
        await updateProfilePage(user);
        
        // Load case history
        await loadCaseHistory();
        
        // Update order page authentication status
        updateOrderPageAuth();
    } else {
        // User not authenticated, showing login
        
        // Desktop: Show login button, hide user dropdown
        if (loginBtn) {
            loginBtn.style.display = 'inline';
        }
        if (userDropdown) {
            userDropdown.classList.add('hidden');
        }

        // Mobile: Show login button, hide user dropdown
        if (loginBtnMobile) {
            loginBtnMobile.style.display = 'inline';
        }
        if (userDropdownMobile) {
            userDropdownMobile.classList.add('hidden');
        }

        // Hide submit menu items for unauthenticated users
        if (submitMenuItem) {
            submitMenuItem.style.display = 'none';
        }
        if (submitMenuItemMobile) {
            submitMenuItemMobile.style.display = 'none';
        }

        // Hide admin menu item for unauthenticated users
        const adminMenuItem = document.getElementById('adminMenuItem');
        if (adminMenuItem) {
            adminMenuItem.style.display = 'none';
        }
        
        // Update order page authentication status
        updateOrderPageAuth();
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
    
    // Load case history
    await loadCaseHistory();
}

// Update user statistics
async function updateUserStats(user) {
    // Get all stat display elements
    const pointsDisplay = document.querySelector('[data-stat="points"]');
    const globalRankDisplay = document.querySelector('[data-stat="global-rank"]');
    const casesDisplay = document.querySelector('[data-stat="cases"]');
    const accuracyDisplay = document.querySelector('[data-stat="accuracy"]');
    const currentRankDisplay = document.querySelector('[data-stat="current-rank"]');
    const currentScoreDisplay = document.querySelector('[data-stat="current-score"]');
    const currentAccuracyDisplay = document.querySelector('[data-stat="current-accuracy"]');

    // Set loading state for global stats
    if (pointsDisplay) pointsDisplay.textContent = '...';
    if (globalRankDisplay) globalRankDisplay.textContent = '...';
    if (casesDisplay) casesDisplay.textContent = '...';
    if (accuracyDisplay) accuracyDisplay.textContent = '...';
    
    // Set loading state for current case stats
    if (currentRankDisplay) currentRankDisplay.textContent = '...';
    if (currentScoreDisplay) currentScoreDisplay.textContent = '...';
    if (currentAccuracyDisplay) currentAccuracyDisplay.textContent = '...';
    
    try {
        // Fetch both leaderboards in parallel
        const [globalResponse, currentResponse] = await Promise.all([
            API.get(API_CONFIG.ENDPOINTS.LEADERBOARD),
            API.get(API_CONFIG.ENDPOINTS.LEADERBOARD_CURRENT).catch(() => ({ leaderboard: [] })) // Don't fail if no current case
        ]);
        
        // Global leaderboard stats
        const globalLeaderboard = globalResponse.leaderboard || [];
        const globalRank = globalLeaderboard.findIndex(entry => entry.user_id === user.id) + 1;
        
        const globalStats = {
            points: user.points || 0,
            rank: globalRank > 0 ? globalRank : '---',
            casesSolved: user.cases_count || 0,
            accuracy: Math.round((user.accuracy || 0) * 100)
        };

        // Update global stats UI
        if (pointsDisplay) pointsDisplay.textContent = globalStats.points.toLocaleString();
        if (globalRankDisplay) globalRankDisplay.textContent = globalStats.rank !== '---' ? `# ${globalStats.rank}` : 'Sin ranking';
        if (casesDisplay) casesDisplay.textContent = globalStats.casesSolved;
        if (accuracyDisplay) accuracyDisplay.textContent = `${globalStats.accuracy}%`;
        
        // Current case leaderboard stats
        const currentLeaderboard = currentResponse.leaderboard || [];
        const currentCaseEntry = currentLeaderboard.find(entry => entry.user_id === user.id);
        
        if (currentCaseEntry) {
            const currentRank = currentLeaderboard.findIndex(entry => entry.user_id === user.id) + 1;
            if (currentRankDisplay) currentRankDisplay.textContent = `# ${currentRank}`;
            if (currentScoreDisplay) currentScoreDisplay.textContent = `${currentCaseEntry.points} pts`;
            if (currentAccuracyDisplay) currentAccuracyDisplay.textContent = `${Math.round(currentCaseEntry.accuracy * 100)}%`;
        } else {
            // User hasn't participated in current case
            if (currentRankDisplay) currentRankDisplay.textContent = 'No participando';
            if (currentScoreDisplay) currentScoreDisplay.textContent = '--';
            if (currentAccuracyDisplay) currentAccuracyDisplay.textContent = '--%';
        }
        
        // Update badges if they exist
        updateUserBadges(user);
        
    } catch (error) {
        console.error('Failed to load user stats:', error);
        
        // Show error state for global stats
        if (pointsDisplay) pointsDisplay.textContent = 'Error';
        if (globalRankDisplay) globalRankDisplay.textContent = 'Error';
        if (casesDisplay) casesDisplay.textContent = 'Error';
        if (accuracyDisplay) accuracyDisplay.textContent = 'Error';
        
        // Show error state for current case stats
        if (currentRankDisplay) currentRankDisplay.textContent = 'Error';
        if (currentScoreDisplay) currentScoreDisplay.textContent = 'Error';
        if (currentAccuracyDisplay) currentAccuracyDisplay.textContent = 'Error';
    }
}

// Update user badges display
function updateUserBadges(user) {
    const badgeContainer = document.getElementById('userBadgesContainer');
    if (!badgeContainer) return;
    
    if (user.badges && user.badges.length > 0) {
        badgeContainer.innerHTML = user.badges.map(badge => 
            `<span class="badge">${badge}</span>`
        ).join('');
    } else {
        badgeContainer.innerHTML = '<span class="badge">🔍 Detective Novato</span>';
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
            // Get full user data from API
            const userResponse = await API.get(API_CONFIG.ENDPOINTS.PROFILE);
            Auth.setUser(userResponse);
        } catch (error) {
            console.error('Failed to get user data:', error);
            // Fallback to token data if API call fails
            try {
                const payload = JSON.parse(atob(token.split('.')[1]));
                const userData = {
                    id: payload.user_id,
                    email: payload.email,
                    name: payload.name
                };
                Auth.setUser(userData);
            } catch (tokenError) {
                console.error('Failed to parse token:', tokenError);
            }
        }

        await updateAuthUI();
        showNotification('¡Sesión iniciada exitosamente!', 'success');
        
        // Clean up the URL
        window.history.replaceState({}, document.title, window.location.pathname);
        
        // Show the profile page
        showPage('profile');
    }
}

// Load case history for user profile
async function loadCaseHistory() {
    const container = document.getElementById('caseHistoryContainer');
    if (!container) return;
    
    try {
        const response = await API.get(`${API_CONFIG.ENDPOINTS.SUBMISSIONS}?limit=10`);
        const submissions = response.submissions || [];
        
        if (submissions.length === 0) {
            container.innerHTML = `
                <div style="text-align: center; opacity: 0.7;">
                    <p>No has completado ningún caso aún.</p>
                    <p>¡Empieza tu primera investigación!</p>
                </div>
            `;
            return;
        }
        
        let html = '';
        submissions.forEach(submission => {
            const submittedDate = new Date(submission.submitted_at).toLocaleDateString('es-ES', {
                year: 'numeric',
                month: 'long',
                day: 'numeric'
            });
            
            const accuracy = Math.round(submission.accuracy * 100);
            
            html += `
                <div style="background: rgba(0,0,0,0.3); padding: 1rem; border-radius: 8px; margin-bottom: 1rem;">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <div>
                            <strong>${submission.case_name}</strong><br>
                            <span style="opacity: 0.8;">Completado: ${submittedDate}</span><br>
                            <span style="opacity: 0.6; font-size: 0.9rem;">Precisión: ${accuracy}%</span>
                        </div>
                        <div style="text-align: right;">
                            <div style="color: #d4af37; font-weight: bold;">${submission.score} pts</div>
                            <div style="color: #2ecc71;">✓ ${submission.status === 'completed' ? 'Resuelto' : 'En progreso'}</div>
                        </div>
                    </div>
                </div>
            `;
        });
        
        container.innerHTML = html;
        
    } catch (error) {
        console.error('Failed to load case history:', error);
        container.innerHTML = `
            <div style="text-align: center; opacity: 0.7;">
                <p>Error al cargar el historial de casos.</p>
            </div>
        `;
    }
}

// Initialize auth UI when page loads
document.addEventListener('DOMContentLoaded', async function() {
    await updateAuthUI();
});

// Debug functions removed for security



