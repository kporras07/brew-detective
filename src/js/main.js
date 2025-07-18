// User dropdown functionality
function toggleUserMenu() {
    const userMenu = document.getElementById('userMenu');
    const userMenuToggle = document.getElementById('userMenuToggle');
    
    userMenu.classList.toggle('active');
    userMenuToggle.classList.toggle('active');
}

function closeUserMenu() {
    const userMenu = document.getElementById('userMenu');
    const userMenuToggle = document.getElementById('userMenuToggle');
    
    userMenu.classList.remove('active');
    userMenuToggle.classList.remove('active');
}

// Close user menu when clicking outside
document.addEventListener('click', function(event) {
    const userDropdown = document.getElementById('userDropdown');
    const userMenu = document.getElementById('userMenu');
    
    if (userDropdown && !userDropdown.contains(event.target)) {
        closeUserMenu();
    }
});

// Mobile menu functionality
function toggleMobileMenu() {
    const mobileNav = document.getElementById('mobileNav');
    const toggle = document.querySelector('.mobile-menu-toggle');
    const body = document.body;
    
    mobileNav.classList.toggle('active');
    toggle.classList.toggle('active');
    
    // Prevent background scrolling when menu is open
    if (mobileNav.classList.contains('active')) {
        body.style.overflow = 'hidden';
    } else {
        body.style.overflow = '';
    }
}

function closeMobileMenu() {
    const mobileNav = document.getElementById('mobileNav');
    const toggle = document.querySelector('.mobile-menu-toggle');
    const body = document.body;
    
    mobileNav.classList.remove('active');
    toggle.classList.remove('active');
    body.style.overflow = '';
}

// Close mobile menu when clicking outside
document.addEventListener('click', function(event) {
    const mobileNav = document.getElementById('mobileNav');
    const toggle = document.querySelector('.mobile-menu-toggle');
    
    if (!toggle.contains(event.target) && !mobileNav.contains(event.target)) {
        closeMobileMenu();
    }
});

// Navigation functionality
function showPage(pageId) {
    // Hide all pages
    const pages = document.querySelectorAll('.page');
    pages.forEach(page => page.classList.remove('active'));
    
    // Show selected page
    document.getElementById(pageId).classList.add('active');
    
    // Load data for specific pages
    if (pageId === 'leaderboard') {
        loadLeaderboard();
    } else if (pageId === 'profile') {
        loadProfile();
    } else if (pageId === 'submit') {
        checkAuthForSubmit();
    }
}

// Check authentication for submit page
function checkAuthForSubmit() {
    if (!Auth.isAuthenticated()) {
        showNotification('Debes iniciar sesión para enviar respuestas', 'error');
        setTimeout(() => {
            showPage('home');
        }, 2000);
        return;
    }
}

// Load profile data
function loadProfile() {
    if (!Auth.isAuthenticated()) {
        showNotification('Debes iniciar sesión para ver tu perfil', 'error');
        setTimeout(() => {
            showPage('home');
        }, 2000);
        return;
    }

    const user = Auth.getUser();
    if (user) {
        updateProfilePage(user);
    }

    // Load real profile data from API
    API.get(API_CONFIG.ENDPOINTS.PROFILE)
        .then(userData => {
            Auth.setUser(userData);
            updateProfilePage(userData);
        })
        .catch(error => {
            console.error('Failed to load profile:', error);
        });
}

// Submit form functionality
document.getElementById('submitForm').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    // Check authentication
    if (!Auth.isAuthenticated()) {
        showNotification('Debes iniciar sesión para enviar respuestas', 'error');
        return;
    }
    
    // Hide previous messages
    document.getElementById('submitSuccess').style.display = 'none';
    document.getElementById('submitError').style.display = 'none';
    
    // Collect form data
    const submission = {
        detective_name: document.getElementById('detectiveName').value,
        case_id: document.getElementById('caseId').value,
        coffee_answers: [
            {
                coffee_id: '1',
                region: document.getElementById('coffee1_region').value,
                variety: document.getElementById('coffee1_variety').value,
                process: document.getElementById('coffee1_process').value,
                roast_level: document.getElementById('coffee1_roast').value,
                tasting_notes: document.getElementById('coffee1_notes').value
            },
            {
                coffee_id: '2',
                region: document.getElementById('coffee2_region').value,
                variety: document.getElementById('coffee2_variety').value,
                process: document.getElementById('coffee2_process').value,
                roast_level: document.getElementById('coffee2_roast').value,
                tasting_notes: document.getElementById('coffee2_notes').value
            },
            {
                coffee_id: '3',
                region: document.getElementById('coffee3_region').value,
                variety: document.getElementById('coffee3_variety').value,
                process: document.getElementById('coffee3_process').value,
                roast_level: document.getElementById('coffee3_roast').value,
                tasting_notes: document.getElementById('coffee3_notes').value
            },
            {
                coffee_id: '4',
                region: document.getElementById('coffee4_region').value,
                variety: document.getElementById('coffee4_variety').value,
                process: document.getElementById('coffee4_process').value,
                roast_level: document.getElementById('coffee4_roast').value,
                tasting_notes: document.getElementById('coffee4_notes').value
            }
        ],
        favorite_coffee: document.getElementById('favorite_coffee').value,
        brewing_method: document.getElementById('brewing_method').value
    };
    
    try {
        const response = await API.post(API_CONFIG.ENDPOINTS.SUBMISSIONS, submission);
        
        // Show success message
        document.getElementById('submitSuccess').style.display = 'block';
        document.getElementById('submitForm').reset();
        
        // Show score if available
        if (response.score) {
            document.getElementById('submitSuccess').innerHTML = 
                `¡Caso resuelto exitosamente! Puntuación: ${response.score} (${Math.round(response.accuracy * 100)}% precisión)`;
        }
        
        // Redirect to leaderboard after delay
        setTimeout(() => {
            showPage('leaderboard');
            loadLeaderboard(); // Refresh leaderboard
        }, 2000);
        
    } catch (error) {
        console.error('Submission failed:', error);
        document.getElementById('submitError').style.display = 'block';
        document.getElementById('submitError').innerHTML = 
            'Error al enviar respuestas. Por favor verifica tu conexión e intenta nuevamente.';
    }
});

// Load leaderboard from API
async function loadLeaderboard() {
    try {
        const response = await API.get(API_CONFIG.ENDPOINTS.LEADERBOARD);
        const leaderboard = response.leaderboard || [];
        
        const leaderboardContainer = document.querySelector('.leaderboard');
        if (!leaderboardContainer) return;
        
        // Clear existing entries
        leaderboardContainer.innerHTML = '';
        
        // Add entries
        leaderboard.forEach((entry, index) => {
            const rank = index + 1;
            const medal = rank === 1 ? '🥇' : rank === 2 ? '🥈' : rank === 3 ? '🥉' : '';
            
            const entryElement = document.createElement('div');
            entryElement.className = 'leaderboard-entry';
            entryElement.innerHTML = `
                <div class="rank">${medal} ${rank}</div>
                <div class="detective-name">
                    ${entry.detective_name || 'Detective Anónimo'}
                    ${entry.badges && entry.badges.length > 0 ? 
                        entry.badges.map(badge => `<span class="badge">${badge}</span>`).join('') : 
                        '<span class="badge">Detective Novato</span>'
                    }
                </div>
                <div class="score">${entry.points || 0} pts</div>
            `;
            
            leaderboardContainer.appendChild(entryElement);
        });
        
        // Add loading success message
        if (leaderboard.length === 0) {
            leaderboardContainer.innerHTML = '<p style="text-align: center; padding: 2rem;">No hay detectives en el ranking aún. ¡Sé el primero!</p>';
        }
        
    } catch (error) {
        console.error('Failed to load leaderboard:', error);
        // Fallback to static leaderboard if API fails
        console.log('Using static leaderboard as fallback');
    }
}

// Profile functionality
async function saveProfile() {
    const name = document.getElementById('profileName').value;
    const email = document.getElementById('profileEmail').value;
    const level = document.getElementById('profileLevel').value;
    
    if (!name || !email) {
        alert('Por favor completa todos los campos obligatorios.');
        return;
    }
    
    try {
        // Generate a simple user ID based on email
        const userId = btoa(email).replace(/[^a-zA-Z0-9]/g, '').substring(0, 20);
        
        const profileData = {
            name: name,
            email: email,
            level: level
        };
        
        await API.put(`${API_CONFIG.ENDPOINTS.USERS}/${userId}`, profileData);
        alert('¡Perfil actualizado exitosamente, Detective ' + name + '!');
        
        // Store user ID in local storage for future use
        localStorage.setItem('brew_detective_user_id', userId);
        
    } catch (error) {
        console.error('Failed to save profile:', error);
        alert('Error al actualizar el perfil. Por favor intenta nuevamente.');
    }
}

// Add some interactive animations
document.addEventListener('DOMContentLoaded', function() {
    // Animate cards on scroll
    const cards = document.querySelectorAll('.card');
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('fade-in');
            }
        });
    });
    
    cards.forEach(card => {
        observer.observe(card);
    });
});

// Simulate dynamic leaderboard updates
setInterval(() => {
    const entries = document.querySelectorAll('.leaderboard-entry');
    if (entries.length > 0) {
        const randomEntry = entries[Math.floor(Math.random() * entries.length)];
        randomEntry.style.background = 'rgba(212, 175, 55, 0.1)';
        setTimeout(() => {
            randomEntry.style.background = 'rgba(0, 0, 0, 0.3)';
        }, 1000);
    }
}, 10000);