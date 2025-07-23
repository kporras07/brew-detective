// Router for handling navigation and URLs
const Router = {
    routes: {
        '/': 'home',
        '/home': 'home',
        '/order': 'order',
        '/submit': 'submit',
        '/leaderboard': 'leaderboard',
        '/ranking': 'leaderboard', // Alias for leaderboard
        '/profile': 'profile',
        '/admin': 'admin',
        '/thankyou': 'thankyou'
    },
    
    init() {
        // Handle initial page load
        this.handleRoute();
        
        // Handle browser back/forward navigation
        window.addEventListener('popstate', () => {
            this.handleRoute();
        });
    },
    
    navigateTo(path) {
        // Update browser URL without page reload
        history.pushState(null, null, path);
        this.handleRoute();
    },
    
    handleRoute() {
        let path = window.location.pathname;
        
        // Handle root path
        if (path === '/' || path === '') {
            path = '/home';
        }
        
        const pageId = this.routes[path] || 'home';
        
        // Show the page without updating URL (since it's already updated)
        this.showPageInternal(pageId);
    },
    
    showPageInternal(pageId) {
        // Hide all pages
        const pages = document.querySelectorAll('.page');
        pages.forEach(page => page.classList.remove('active'));
        
        // Show selected page
        const targetPage = document.getElementById(pageId);
        if (targetPage) {
            targetPage.classList.add('active');
        } else {
            // Fallback to home if page doesn't exist
            document.getElementById('home').classList.add('active');
        }
        
        // Load data for specific pages
        if (pageId === 'leaderboard') {
            loadLeaderboard();
        } else if (pageId === 'profile') {
            loadProfile();
        } else if (pageId === 'submit') {
            checkAuthForSubmit();
            loadActiveCase();
            loadCatalogData();
        } else if (pageId === 'admin') {
            checkAdminAccess();
            loadCaseFormDropdowns();
            loadOrderFormDropdowns();
        } else if (pageId === 'order') {
            updateOrderPageAuth();
        }
        
        // Update page title
        this.updatePageTitle(pageId);
    },
    
    updatePageTitle(pageId) {
        const titles = {
            'home': 'Brew Detective - Misterio Cafetalero de Costa Rica',
            'order': 'Ordenar Caja - Brew Detective',
            'submit': 'Enviar Respuestas - Brew Detective', 
            'leaderboard': 'Ranking de Detectives - Brew Detective',
            'profile': 'Mi Perfil - Brew Detective',
            'admin': 'Administración - Brew Detective'
        };
        
        document.title = titles[pageId] || titles['home'];
    }
};

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
    // Find the path for this page
    const path = Object.keys(Router.routes).find(key => Router.routes[key] === pageId) || `/${pageId}`;
    
    // Use router to navigate
    Router.navigateTo(path);
}

// Setup navigation links to work with router
function setupNavigationLinks() {
    // Handle all navigation links
    document.addEventListener('click', function(event) {
        const link = event.target.closest('a[href^="/"]');
        if (link) {
            // Prevent default page reload
            event.preventDefault();
            
            // Get the href and navigate using router
            const href = link.getAttribute('href');
            Router.navigateTo(href);
        }
    });
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

// Update order page based on authentication status
function updateOrderPageAuth() {
    const authenticatedContent = document.getElementById('orderAuthenticatedContent');
    const unauthenticatedContent = document.getElementById('orderUnauthenticatedContent');
    
    if (Auth.isAuthenticated()) {
        // User is logged in - show WhatsApp button
        authenticatedContent.style.display = 'block';
        unauthenticatedContent.style.display = 'none';
    } else {
        // User is not logged in - show login prompt
        authenticatedContent.style.display = 'none';
        unauthenticatedContent.style.display = 'block';
    }
}

// Open WhatsApp order with user email included
function openWhatsAppOrder() {
    const user = Auth.getUser();
    
    if (!user) {
        showNotification('Error: No se pudo obtener la información del usuario', 'error');
        return;
    }
    
    const userName = user.name || 'Usuario';
    const userEmail = user.email || '';
    
    // Create the WhatsApp message with user information
    const message = `Hola! Quiero ordenar una caja Brew Detective por ₡18,000

Nombre: ${userName}
Email: ${userEmail}

Por favor coordinen conmigo la entrega y el pago. ¡Gracias!`;
    
    // Encode the message for URL
    const encodedMessage = encodeURIComponent(message);
    
    // WhatsApp phone number
    const phoneNumber = '50685489236';
    
    // Create the WhatsApp URL
    const whatsappUrl = `https://wa.me/${phoneNumber}?text=${encodedMessage}`;
    
    // Open WhatsApp in a new window/tab
    window.open(whatsappUrl, '_blank', 'noopener,noreferrer');
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
            // Load case history
            if (typeof loadCaseHistory === 'function') {
                loadCaseHistory();
            }
        })
        .catch(error => {
            console.error('Failed to load profile:', error);
        });
}

// Sanitize case data to remove sensitive information (answers)
function sanitizeCaseData(caseData) {
    if (!caseData) return null;
    
    // Create a safe version with only public information
    const safeCaseData = {
        id: caseData.id,
        name: caseData.name,
        description: caseData.description,
        is_active: caseData.is_active,
        enabled_questions: caseData.enabled_questions,
        // Only include coffee IDs, not the full coffee objects with answers
        coffeeIds: caseData.coffees ? caseData.coffees.map(coffee => coffee.id) : [],
        coffee_count: caseData.coffees ? caseData.coffees.length : 0
    };
    
    return safeCaseData;
}

// Load active case information
async function loadActiveCase() {
    const activeCaseName = document.getElementById('activeCaseName');
    const activeCaseDescription = document.getElementById('activeCaseDescription');
    
    if (!activeCaseName) return;
    
    try {
        // Show loading state
        activeCaseName.textContent = 'Cargando caso activo...';
        activeCaseDescription.textContent = '';
        
        // Try secure endpoint first, fallback to regular endpoint with sanitization
        let response, activeCase;
        
        try {
            // Try the secure public endpoint (no answers included)
            response = await API.get(API_CONFIG.ENDPOINTS.ACTIVE_CASE_PUBLIC);
            activeCase = response.case;
        } catch (error) {
            console.warn('Secure endpoint not available, using fallback with sanitization');
            // Fallback to regular endpoint but sanitize the response
            response = await API.get(API_CONFIG.ENDPOINTS.ACTIVE_CASE);
            activeCase = sanitizeCaseData(response.case);
        }
        
        if (activeCase) {
            activeCaseName.textContent = activeCase.name;
            activeCaseDescription.textContent = activeCase.description;
            
            // Store the already-sanitized case data globally
            window.activeCaseData = {
                id: activeCase.id,
                name: activeCase.name,
                description: activeCase.description,
                // Handle both camelCase and snake_case from different endpoints
                coffeeIds: activeCase.coffeeIds || activeCase.coffee_ids || [],
                questions: activeCase.enabled_questions || {
                    region: true,
                    variety: true,
                    process: true,
                    taste_note_1: true,
                    taste_note_2: true,
                    favorite_coffee: true,
                    brewing_method: true
                }
            };
            
            // Explicitly remove any potentially unsafe data
            delete window.activeCase;
            delete window.activeCaseQuestions;
            if (response.case && response.case.coffees) {
                delete response.case.coffees; // Remove coffee details if they exist
            }
            
            // Case loaded successfully
            
            // Customize submission form based on enabled questions
            customizeSubmissionForm();
            
            // Update scoring information box
            updateScoringInfo();
        } else {
            activeCaseName.textContent = 'No hay caso activo';
            activeCaseDescription.textContent = 'Por favor contacta al administrador.';
        }
    } catch (error) {
        console.error('Failed to load active case:', error);
        activeCaseName.textContent = 'Error al cargar caso';
        activeCaseDescription.textContent = 'No se pudo obtener la información del caso activo.';
    }
}

// Customize submission form based on active case enabled questions
function customizeSubmissionForm() {
    const questions = window.activeCaseQuestions;
    if (!questions) return;
    
    // Helper function to show/hide form groups for coffee questions
    function toggleCoffeeQuestion(questionType, enabled) {
        for (let i = 1; i <= 4; i++) {
            const formGroup = document.querySelector(`#coffee${i}_${questionType}`);
            if (formGroup && formGroup.closest('.form-group')) {
                formGroup.closest('.form-group').style.display = enabled ? 'block' : 'none';
                
                // Make required/optional based on enabled state
                formGroup.required = enabled;
            }
        }
    }
    
    // Show/hide coffee-specific questions
    toggleCoffeeQuestion('region', questions.region);
    toggleCoffeeQuestion('variety', questions.variety);
    toggleCoffeeQuestion('process', questions.process);
    
    // Handle taste notes
    for (let i = 1; i <= 4; i++) {
        const note1Input = document.getElementById(`coffee${i}_note1`);
        const note2Input = document.getElementById(`coffee${i}_note2`);
        
        if (note1Input && note1Input.closest('.form-group')) {
            note1Input.closest('.form-group').style.display = questions.taste_note_1 ? 'block' : 'none';
        }
        if (note2Input && note2Input.closest('.form-group')) {
            note2Input.closest('.form-group').style.display = questions.taste_note_2 ? 'block' : 'none';
        }
    }
    
    // Show/hide bonus questions
    const favoriteCoffeeGroup = document.getElementById('favorite_coffee');
    if (favoriteCoffeeGroup && favoriteCoffeeGroup.closest('.form-group')) {
        favoriteCoffeeGroup.closest('.form-group').style.display = questions.favorite_coffee ? 'block' : 'none';
    }
    
    const brewingMethodGroup = document.getElementById('brewing_method');
    if (brewingMethodGroup && brewingMethodGroup.closest('.form-group')) {
        brewingMethodGroup.closest('.form-group').style.display = questions.brewing_method ? 'block' : 'none';
    }
    
    // Hide entire bonus questions section if no bonus questions are enabled
    const bonusSection = document.querySelector('h3[style*="🔍 Preguntas Bonus"]');
    if (bonusSection) {
        const showBonusSection = questions.favorite_coffee || questions.brewing_method;
        bonusSection.style.display = showBonusSection ? 'block' : 'none';
    }
}

// Update scoring information box based on enabled questions
function updateScoringInfo() {
    const questions = window.activeCaseData?.questions || {};
    if (!questions || Object.keys(questions).length === 0) return;
    
    const coffeeQuestionsList = document.getElementById('coffeeQuestionsList');
    const bonusQuestionsList = document.getElementById('bonusQuestionsList');
    const maxScoreDisplay = document.getElementById('maxScoreDisplay');
    
    if (!coffeeQuestionsList || !bonusQuestionsList || !maxScoreDisplay) return;
    
    // Count enabled coffee questions
    let enabledCoffeeQuestions = 0;
    const coffeeQuestionsHtml = [];
    
    if (questions.region) {
        coffeeQuestionsHtml.push('<div>• Región: <strong>20 puntos</strong></div>');
        enabledCoffeeQuestions++;
    }
    if (questions.variety) {
        coffeeQuestionsHtml.push('<div>• Variedad: <strong>20 puntos</strong></div>');
        enabledCoffeeQuestions++;
    }
    if (questions.process) {
        coffeeQuestionsHtml.push('<div>• Proceso: <strong>20 puntos</strong></div>');
        enabledCoffeeQuestions++;
    }
    if (questions.taste_note_1) {
        coffeeQuestionsHtml.push('<div>• Nota de cata 1: <strong>20 puntos</strong></div>');
        enabledCoffeeQuestions++;
    }
    if (questions.taste_note_2) {
        coffeeQuestionsHtml.push('<div>• Nota de cata 2: <strong>20 puntos</strong></div>');
        enabledCoffeeQuestions++;
    }
    
    // Update coffee questions display
    if (enabledCoffeeQuestions > 0) {
        const pointsPerQuestion = Math.round(100 / enabledCoffeeQuestions);
        const adjustedHtml = coffeeQuestionsHtml.map(html => 
            html.replace('20 puntos', `${pointsPerQuestion} puntos`)
        );
        coffeeQuestionsList.innerHTML = adjustedHtml.join('');
    } else {
        coffeeQuestionsList.innerHTML = '<div style="opacity: 0.7; font-style: italic;">No hay preguntas de café habilitadas</div>';
    }
    
    // Count enabled bonus questions and update display
    const bonusQuestionsHtml = [];
    let bonusPoints = 0;
    
    if (questions.favorite_coffee) {
        bonusQuestionsHtml.push('<div>• Café favorito: <strong>+50 puntos</strong></div>');
        bonusPoints += 50;
    }
    if (questions.brewing_method) {
        bonusQuestionsHtml.push('<div>• Método de preparación: <strong>+50 puntos</strong></div>');
        bonusPoints += 50;
    }
    
    if (bonusQuestionsHtml.length > 0) {
        bonusQuestionsList.innerHTML = bonusQuestionsHtml.join('');
    } else {
        bonusQuestionsList.innerHTML = '<div style="opacity: 0.7; font-style: italic;">No hay preguntas bonus habilitadas</div>';
    }
    
    // Calculate and display maximum score
    // Formula: (base points × 4 coffees) + bonus points
    const basePointsPerCoffee = enabledCoffeeQuestions > 0 ? 100 : 0;
    const maxBaseScore = basePointsPerCoffee * 4; // 4 coffees
    const maxTotalScore = maxBaseScore + bonusPoints;
    
    maxScoreDisplay.innerHTML = `<strong style="color: #d4af37; font-size: 1rem;">🏆 Máximo: ${maxTotalScore} puntos</strong>`;
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
    
    // Get coffee IDs from active case data
    const activeCaseData = window.activeCaseData;
    if (!activeCaseData || !activeCaseData.coffeeIds || activeCaseData.coffeeIds.length < 4) {
        // Debug information for troubleshooting
        console.error('Submit error - Active case data:', activeCaseData);
        console.error('Coffee IDs available:', activeCaseData?.coffeeIds);
        
        document.getElementById('submitError').style.display = 'block';
        document.getElementById('submitError').innerHTML = 'Error: No se pudo obtener la información del caso activo. Por favor recarga la página.';
        return;
    }
    
    // Security check: Ensure no coffee answer data is accessible
    if (window.activeCase || window.activeCaseQuestions) {
        delete window.activeCase;
        delete window.activeCaseQuestions;
        console.warn('Removed potentially unsafe case data from global scope');
    }
    
    // Collect form data with actual coffee UUIDs
    const submission = {
        order_id: document.getElementById('orderId').value.toUpperCase(),
        coffee_answers: [
            {
                coffee_id: activeCaseData.coffeeIds[0],
                region: document.getElementById('coffee1_region').value,
                variety: document.getElementById('coffee1_variety').value,
                process: document.getElementById('coffee1_process').value,
                taste_note_1: document.getElementById('coffee1_note1').value,
                taste_note_2: document.getElementById('coffee1_note2').value
            },
            {
                coffee_id: activeCaseData.coffeeIds[1],
                region: document.getElementById('coffee2_region').value,
                variety: document.getElementById('coffee2_variety').value,
                process: document.getElementById('coffee2_process').value,
                taste_note_1: document.getElementById('coffee2_note1').value,
                taste_note_2: document.getElementById('coffee2_note2').value
            },
            {
                coffee_id: activeCaseData.coffeeIds[2],
                region: document.getElementById('coffee3_region').value,
                variety: document.getElementById('coffee3_variety').value,
                process: document.getElementById('coffee3_process').value,
                taste_note_1: document.getElementById('coffee3_note1').value,
                taste_note_2: document.getElementById('coffee3_note2').value
            },
            {
                coffee_id: activeCaseData.coffeeIds[3],
                region: document.getElementById('coffee4_region').value,
                variety: document.getElementById('coffee4_variety').value,
                process: document.getElementById('coffee4_process').value,
                taste_note_1: document.getElementById('coffee4_note1').value,
                taste_note_2: document.getElementById('coffee4_note2').value
            }
        ],
        favorite_coffee: document.getElementById('favorite_coffee').value,
        brewing_method: document.getElementById('brewing_method').value
    };
    
    // Submitting case with coffee answers
    
    try {
        const response = await API.post(API_CONFIG.ENDPOINTS.SUBMISSIONS, submission);
        
        // Store score data for thank you page
        if (response.score !== undefined && response.accuracy !== undefined) {
            sessionStorage.setItem('lastSubmissionScore', response.score);
            sessionStorage.setItem('lastSubmissionAccuracy', response.accuracy);
        }
        
        // Show success message briefly
        document.getElementById('submitSuccess').style.display = 'block';
        document.getElementById('submitForm').reset();
        
        // Redirect to thank you page immediately
        setTimeout(() => {
            showPage('thankyou');
            populateThankYouPage(response);
        }, 1000);
        
    } catch (error) {
        console.error('Submission failed:', error);
        document.getElementById('submitError').style.display = 'block';
        
        // Use the detailed error message from the backend if available
        let errorMessage = error.message;
        
        // If it's a generic HTTP error or connection issue, provide fallback
        if (!errorMessage || errorMessage.includes('HTTP error! status:') || errorMessage.includes('Failed to fetch')) {
            errorMessage = 'Error al enviar respuestas. Por favor verifica tu conexión e intenta nuevamente.';
        }
        
        document.getElementById('submitError').innerHTML = errorMessage;
    }
});

// Populate thank you page with score data
function populateThankYouPage(submissionResponse) {
    const finalScoreElement = document.getElementById('finalScore');
    const finalAccuracyElement = document.getElementById('finalAccuracy');
    
    if (finalScoreElement && finalAccuracyElement) {
        // Use response data if available, otherwise try sessionStorage
        const score = submissionResponse?.score ?? sessionStorage.getItem('lastSubmissionScore') ?? '--';
        const accuracy = submissionResponse?.accuracy ?? sessionStorage.getItem('lastSubmissionAccuracy') ?? 0;
        
        finalScoreElement.textContent = score;
        finalAccuracyElement.textContent = accuracy !== '--' ? `${Math.round(accuracy * 100)}%` : '--%';
    }
}

// Load both leaderboards from API
async function loadLeaderboard() {
    await Promise.all([
        loadCurrentCaseLeaderboard(),
        loadGlobalLeaderboard()
    ]);
}

// Load current case leaderboard
async function loadCurrentCaseLeaderboard() {
    const container = document.getElementById('currentCaseLeaderboard');
    const infoContainer = document.getElementById('currentCaseInfo');
    if (!container) return;
    
    // Show loading spinner
    container.innerHTML = `
        <div class="loading-spinner">
            <div class="spinner"></div>
            <p>Cargando ranking del caso actual...</p>
        </div>
    `;
    
    try {
        const response = await API.get(API_CONFIG.ENDPOINTS.LEADERBOARD_CURRENT);
        const leaderboard = response.leaderboard || [];
        
        // Update case info
        if (infoContainer && response.case_name) {
            infoContainer.innerHTML = `<p><strong>${response.case_name}</strong> - ${leaderboard.length} detectives participando</p>`;
        }
        
        // Clear loading spinner
        container.innerHTML = '';
        
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
                    <!-- Badges temporarily disabled -->
                </div>
                <div class="score">${entry.points || 0} pts (${Math.round(entry.accuracy * 100)}%)</div>
            `;
            
            container.appendChild(entryElement);
        });
        
        // Show empty state if no data
        if (leaderboard.length === 0) {
            container.innerHTML = '<p style="text-align: center; padding: 2rem;">No hay detectives en este caso aún. ¡Sé el primero en resolverlo!</p>';
            if (infoContainer) {
                infoContainer.innerHTML = '<p>No hay participantes en el caso actual</p>';
            }
        }
        
    } catch (error) {
        console.error('Failed to load current case leaderboard:', error);
        container.innerHTML = '<p style="text-align: center; padding: 2rem; color: #e74c3c;">Error al cargar el ranking del caso actual.</p>';
        if (infoContainer) {
            infoContainer.innerHTML = '<p style="color: #e74c3c;">Error al cargar información del caso</p>';
        }
    }
}

// Load global/historical leaderboard
async function loadGlobalLeaderboard() {
    const container = document.getElementById('globalLeaderboard');
    if (!container) return;
    
    // Show loading spinner
    container.innerHTML = `
        <div class="loading-spinner">
            <div class="spinner"></div>
            <p>Cargando ranking global...</p>
        </div>
    `;
    
    try {
        const response = await API.get(API_CONFIG.ENDPOINTS.LEADERBOARD);
        const leaderboard = response.leaderboard || [];
        
        // Clear loading spinner
        container.innerHTML = '';
        
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
                    <!-- Badges temporarily disabled -->
                </div>
                <div class="score">${entry.points || 0} pts (${entry.cases_count || 0} casos)</div>
            `;
            
            container.appendChild(entryElement);
        });
        
        // Show empty state if no data
        if (leaderboard.length === 0) {
            container.innerHTML = '<p style="text-align: center; padding: 2rem;">No hay detectives en el ranking global aún. ¡Sé el primero!</p>';
        }
        
    } catch (error) {
        console.error('Failed to load global leaderboard:', error);
        container.innerHTML = '<p style="text-align: center; padding: 2rem; color: #e74c3c;">Error al cargar el ranking global.</p>';
    }
}

// Profile functionality
async function saveProfile() {
    const name = document.getElementById('profileName').value;
    const email = document.getElementById('profileEmail').value;
    
    if (!name) {
        alert('Por favor completa el nombre.');
        return;
    }
    
    // Check if user is authenticated
    if (!Auth.isAuthenticated()) {
        alert('Debes iniciar sesión para actualizar tu perfil.');
        return;
    }
    
    const user = Auth.getUser();
    if (!user || !user.id) {
        alert('Error: No se pudo obtener la información del usuario.');
        return;
    }
    
    try {
        const profileData = {
            name: name
        };
        
        await API.put(`${API_CONFIG.ENDPOINTS.USERS}/${user.id}`, profileData);
        alert('¡Perfil actualizado exitosamente, Detective ' + name + '!');
        
        // Update local user data
        const updatedUser = { ...user, name: name };
        Auth.setUser(updatedUser);
        
        // Update UI
        await updateAuthUI();
        
    } catch (error) {
        console.error('Failed to save profile:', error);
        alert('Error al actualizar el perfil. Por favor intenta nuevamente.');
    }
}

// Load catalog data for dropdowns
async function loadCatalogData() {
    try {
        const data = await API.get(API_CONFIG.ENDPOINTS.CATALOG);
        const catalog = data.catalog;
        
        // Populate region dropdowns
        populateDropdown('region', catalog.region || []);
        
        // Populate variety dropdowns  
        populateDropdown('variety', catalog.variety || []);
        
        // Populate process dropdowns
        populateDropdown('process', catalog.process || []);
        
        // Populate brewing method dropdown
        populateBrewingMethodDropdown(catalog.brewing_method || []);
        
    } catch (error) {
        console.error('Failed to load catalog data:', error);
        // Keep the existing hardcoded options as fallback
    }
}

// Helper function to populate dropdown options
function populateDropdown(category, items) {
    // Find all dropdowns for this category (coffee1_region, coffee2_region, etc.)
    const selects = document.querySelectorAll(`select[id*="_${category}"]`);
    
    selects.forEach(select => {
        // Keep the first "Select" option
        const firstOption = select.firstElementChild;
        select.innerHTML = '';
        select.appendChild(firstOption);
        
        // Add catalog items
        items.forEach(item => {
            const option = document.createElement('option');
            option.value = item.value;
            option.textContent = item.label;
            select.appendChild(option);
        });
    });
}

// Helper function to populate brewing method dropdown
function populateBrewingMethodDropdown(items) {
    const select = document.getElementById('brewing_method');
    if (!select) return;
    
    // Keep the first "Select" option
    const firstOption = select.firstElementChild;
    select.innerHTML = '';
    select.appendChild(firstOption);
    
    // Add catalog items
    items.forEach(item => {
        const option = document.createElement('option');
        option.value = item.value;
        option.textContent = item.label;
        select.appendChild(option);
    });
}

// Admin functions
function checkAdminAccess() {
    if (!Auth.isAuthenticated()) {
        showNotification('Debes iniciar sesión para acceder al panel de administración', 'error');
        setTimeout(() => {
            showPage('home');
        }, 2000);
        return;
    }
    
    const user = Auth.getUser();
    if (!user || user.type !== 'admin') {
        showNotification('Acceso denegado: Se requieren privilegios de administrador', 'error');
        setTimeout(() => {
            showPage('home');
        }, 2000);
        return;
    }
}

function loadCatalogItems() {
    const category = document.getElementById('adminCategorySelect').value;
    
    if (!category) {
        document.getElementById('catalogItemsList').innerHTML = '';
        document.getElementById('addItemBtn').classList.add('hidden');
        document.getElementById('refreshBtn').classList.add('hidden');
        return;
    }
    
    document.getElementById('addItemBtn').classList.remove('hidden');
    document.getElementById('refreshBtn').classList.remove('hidden');
    
    // Show loading
    document.getElementById('catalogItemsList').innerHTML = '<p>Cargando items...</p>';
    
    API.get(`${API_CONFIG.ENDPOINTS.ADMIN_CATALOG}?category=${category}`)
        .then(data => {
            displayCatalogItems(data.items || []);
        })
        .catch(error => {
            console.error('Error loading catalog items:', error);
            showAdminNotification('Error al cargar los items del catálogo', 'error');
            document.getElementById('catalogItemsList').innerHTML = '<p>Error al cargar los items</p>';
        });
}

function displayCatalogItems(items) {
    const container = document.getElementById('catalogItemsList');
    
    if (items.length === 0) {
        container.innerHTML = '<p>No hay items en esta categoría</p>';
        return;
    }
    
    let html = '<div style="display: grid; gap: 1rem;">';
    
    items.forEach(item => {
        html += `
            <div style="background: rgba(0,0,0,0.3); padding: 1rem; border-radius: 8px; display: flex; justify-content: space-between; align-items: center; position: relative; z-index: 5;">
                <div>
                    <strong>${item.label}</strong> (${item.value})
                    <br>
                    <small>Orden: ${item.display_order} | ${item.is_active ? 'Activo' : 'Inactivo'}</small>
                </div>
                <div>
                    <button onclick="editCatalogItem('${item.id}')" class="cta-button" style="background: #d4af37; margin-right: 0.5rem; padding: 0.5rem 1rem; position: relative; z-index: 10; pointer-events: auto;">
                        Editar
                    </button>
                    <button onclick="deleteCatalogItem('${item.id}')" class="cta-button" style="background: #e74c3c; padding: 0.5rem 1rem; position: relative; z-index: 10; pointer-events: auto;">
                        Eliminar
                    </button>
                </div>
            </div>
        `;
    });
    
    html += '</div>';
    container.innerHTML = html;
}

function showAddItemForm() {
    document.getElementById('addItemForm').classList.add('admin-section__form--visible');
    document.getElementById('newItemValue').value = '';
    document.getElementById('newItemLabel').value = '';
    document.getElementById('newItemOrder').value = '0';
    document.getElementById('newItemActive').checked = true;
}

function cancelAddItem() {
    document.getElementById('addItemForm').classList.remove('admin-section__form--visible');
}

function createCatalogItem() {
    const category = document.getElementById('adminCategorySelect').value;
    const value = document.getElementById('newItemValue').value.trim();
    const label = document.getElementById('newItemLabel').value.trim();
    const order = parseInt(document.getElementById('newItemOrder').value) || 0;
    const isActive = document.getElementById('newItemActive').checked;
    
    if (!value || !label) {
        showAdminNotification('Valor y etiqueta son requeridos', 'error');
        return;
    }
    
    const newItem = {
        category: category,
        value: value,
        label: label,
        display_order: order,
        is_active: isActive
    };
    
    API.post(API_CONFIG.ENDPOINTS.ADMIN_CATALOG, newItem)
        .then(data => {
            showAdminNotification('Item creado exitosamente', 'success');
            cancelAddItem();
            loadCatalogItems();
        })
        .catch(error => {
            console.error('Error creating catalog item:', error);
            showAdminNotification('Error al crear el item', 'error');
        });
}

function editCatalogItem(itemId) {
    // For now, just show a simple prompt
    const newLabel = prompt('Nuevo nombre para mostrar:');
    if (newLabel && newLabel.trim()) {
        const updates = { label: newLabel.trim() };
        
        API.put(`${API_CONFIG.ENDPOINTS.ADMIN_CATALOG}/${itemId}`, updates)
            .then(() => {
                showAdminNotification('Item actualizado exitosamente', 'success');
                loadCatalogItems();
            })
            .catch(error => {
                console.error('Error updating catalog item:', error);
                showAdminNotification('Error al actualizar el item', 'error');
            });
    }
}

function deleteCatalogItem(itemId) {
    if (confirm('¿Estás seguro de que quieres eliminar este item?')) {
        API.delete(`${API_CONFIG.ENDPOINTS.ADMIN_CATALOG}/${itemId}`)
            .then(() => {
                showAdminNotification('Item eliminado exitosamente', 'success');
                loadCatalogItems();
            })
            .catch(error => {
                console.error('Error deleting catalog item:', error);
                showAdminNotification('Error al eliminar el item', 'error');
            });
    }
}

function showAdminNotification(message, type) {
    const successElement = document.getElementById('adminSuccess');
    const errorElement = document.getElementById('adminError');
    
    // Hide both first
    successElement.style.display = 'none';
    errorElement.style.display = 'none';
    
    // Show the appropriate one
    if (type === 'success') {
        successElement.textContent = message;
        successElement.style.display = 'block';
        setTimeout(() => {
            successElement.style.display = 'none';
        }, 3000);
    } else {
        errorElement.textContent = message;
        errorElement.style.display = 'block';
        setTimeout(() => {
            errorElement.style.display = 'none';
        }, 3000);
    }
}

// Case Management Functions

function showCreateCaseForm() {
    document.getElementById('createCaseForm').classList.add('admin-section__form--visible');
    clearCaseForm();
}

function cancelCreateCase() {
    document.getElementById('createCaseForm').classList.remove('admin-section__form--visible');
}

function clearCaseForm() {
    // Clear all form fields
    document.getElementById('caseName').value = '';
    document.getElementById('caseDescription').value = '';
    document.getElementById('caseIsActive').checked = false;
    
    // Clear coffee details
    for (let i = 1; i <= 4; i++) {
        document.getElementById(`coffee${i}Name`).value = '';
        document.getElementById(`coffee${i}Region`).value = '';
        document.getElementById(`coffee${i}Variety`).value = '';
        document.getElementById(`coffee${i}Process`).value = '';
        document.getElementById(`coffee${i}Notes`).value = '';
    }
    
    // Reset question checkboxes to default (all checked)
    document.getElementById('questionRegion').checked = true;
    document.getElementById('questionVariety').checked = true;
    document.getElementById('questionProcess').checked = true;
    document.getElementById('questionTasteNote1').checked = true;
    document.getElementById('questionTasteNote2').checked = true;
    document.getElementById('questionFavoriteCoffee').checked = true;
    document.getElementById('questionBrewingMethod').checked = true;
}

async function loadCaseFormDropdowns() {
    try {
        const data = await API.get(API_CONFIG.ENDPOINTS.CATALOG);
        const catalog = data.catalog;
        
        // Catalog data loaded successfully
        
        // Populate dropdowns for each coffee
        for (let i = 1; i <= 4; i++) {
            populateCaseDropdown(`coffee${i}Region`, catalog.region || []);
            populateCaseDropdown(`coffee${i}Variety`, catalog.variety || []);
            populateCaseDropdown(`coffee${i}Process`, catalog.process || []);
        }
        
        // Also populate edit form dropdowns if they exist
        for (let i = 1; i <= 4; i++) {
            populateCaseDropdown(`editCoffee${i}Region`, catalog.region || []);
            populateCaseDropdown(`editCoffee${i}Variety`, catalog.variety || []);
            populateCaseDropdown(`editCoffee${i}Process`, catalog.process || []);
        }
        
    } catch (error) {
        console.error('Failed to load catalog data for case form:', error);
        showAdminNotification('Error al cargar datos del catálogo. Verifica que existan elementos en el catálogo.', 'error');
    }
}

function populateCaseDropdown(selectId, items) {
    const select = document.getElementById(selectId);
    if (!select) return;
    
    // Keep the first "Select" option
    const firstOption = select.firstElementChild;
    select.innerHTML = '';
    select.appendChild(firstOption);
    
    // Add catalog items
    items.forEach(item => {
        const option = document.createElement('option');
        option.value = item.value;
        option.textContent = item.label;
        select.appendChild(option);
    });
}

async function createCase() {
    const name = document.getElementById('caseName').value.trim();
    const description = document.getElementById('caseDescription').value.trim();
    const isActive = document.getElementById('caseIsActive').checked;
    
    if (!name || !description) {
        showAdminNotification('Nombre y descripción son requeridos', 'error');
        return;
    }
    
    // Collect coffee data
    const coffees = [];
    for (let i = 1; i <= 4; i++) {
        const coffeeName = document.getElementById(`coffee${i}Name`).value.trim();
        const region = document.getElementById(`coffee${i}Region`).value;
        const variety = document.getElementById(`coffee${i}Variety`).value;
        const process = document.getElementById(`coffee${i}Process`).value;
        const notes = document.getElementById(`coffee${i}Notes`).value.trim();
        
        if (!coffeeName || !region || !variety || !process || !notes) {
            showAdminNotification(`Todos los campos del Café #${i} son requeridos`, 'error');
            return;
        }
        
        coffees.push({
            id: `coffee_${i}`,
            name: coffeeName,
            region: region,
            variety: variety,
            process: process,
            tasting_notes: notes
        });
    }
    
    // Collect enabled questions
    const enabledQuestions = {
        region: document.getElementById('questionRegion').checked,
        variety: document.getElementById('questionVariety').checked,
        process: document.getElementById('questionProcess').checked,
        taste_note_1: document.getElementById('questionTasteNote1').checked,
        taste_note_2: document.getElementById('questionTasteNote2').checked,
        favorite_coffee: document.getElementById('questionFavoriteCoffee').checked,
        brewing_method: document.getElementById('questionBrewingMethod').checked
    };

    const newCase = {
        name: name,
        description: description,
        is_active: isActive,
        coffees: coffees,
        enabled_questions: enabledQuestions
    };
    
    try {
        await API.post(API_CONFIG.ENDPOINTS.ADMIN_CASES, newCase);
        showAdminNotification('Caso creado exitosamente', 'success');
        cancelCreateCase();
        loadAllCases();
    } catch (error) {
        console.error('Error creating case:', error);
        showAdminNotification('Error al crear el caso', 'error');
    }
}

async function loadAllCases() {
    const container = document.getElementById('casesList');
    
    // Show loading
    container.innerHTML = '<p>Cargando casos...</p>';
    
    try {
        const data = await API.get(`${API_CONFIG.ENDPOINTS.ADMIN_CASES}?limit=20`);
        displayCases(data.cases || []);
    } catch (error) {
        console.error('Error loading cases:', error);
        showAdminNotification('Error al cargar los casos', 'error');
        container.innerHTML = '<p>Error al cargar los casos</p>';
    }
}

function displayCases(cases) {
    const container = document.getElementById('casesList');
    
    if (cases.length === 0) {
        container.innerHTML = '<p>No hay casos creados</p>';
        return;
    }
    
    let html = '<div style="display: grid; gap: 1rem;">';
    
    cases.forEach(caseItem => {
        const createdDate = new Date(caseItem.created_at).toLocaleDateString('es-ES');
        const coffeeCount = caseItem.coffees ? caseItem.coffees.length : 0;
        
        html += `
            <div style="background: rgba(0,0,0,0.3); padding: 1rem; border-radius: 8px; position: relative; z-index: 5;">
                <div style="display: flex; justify-content: space-between; align-items: start;">
                    <div style="flex: 1;">
                        <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem;">
                            <strong style="color: #d4af37;">${caseItem.name}</strong>
                            ${caseItem.is_active ? 
                                '<span style="background: #2ecc71; color: white; padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.8rem;">ACTIVO</span>' : 
                                '<span style="background: #666; color: white; padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.8rem;">INACTIVO</span>'
                            }
                        </div>
                        <p style="margin: 0.5rem 0; opacity: 0.8;">${caseItem.description}</p>
                        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 0.5rem; margin-top: 1rem;">
                            ${caseItem.coffees ? caseItem.coffees.map(coffee => `
                                <div style="background: rgba(212, 175, 55, 0.1); padding: 0.5rem; border-radius: 4px; font-size: 0.9rem;">
                                    <strong>${coffee.name}</strong><br>
                                    <span style="opacity: 0.8;">${coffee.region} • ${coffee.variety} • ${coffee.process}</span>
                                </div>
                            `).join('') : '<span style="opacity: 0.6;">Sin información de cafés</span>'}
                        </div>
                        ${caseItem.enabled_questions ? `
                            <div style="margin-top: 1rem;">
                                <strong style="color: #d4af37; font-size: 0.9rem;">Preguntas Habilitadas:</strong><br>
                                <div style="display: flex; flex-wrap: wrap; gap: 0.3rem; margin-top: 0.5rem;">
                                    ${caseItem.enabled_questions.region ? '<span style="background: #2ecc71; color: white; padding: 0.2rem 0.4rem; border-radius: 3px; font-size: 0.8rem;">Región</span>' : ''}
                                    ${caseItem.enabled_questions.variety ? '<span style="background: #2ecc71; color: white; padding: 0.2rem 0.4rem; border-radius: 3px; font-size: 0.8rem;">Variedad</span>' : ''}
                                    ${caseItem.enabled_questions.process ? '<span style="background: #2ecc71; color: white; padding: 0.2rem 0.4rem; border-radius: 3px; font-size: 0.8rem;">Proceso</span>' : ''}
                                    ${caseItem.enabled_questions.taste_note_1 ? '<span style="background: #2ecc71; color: white; padding: 0.2rem 0.4rem; border-radius: 3px; font-size: 0.8rem;">Nota 1</span>' : ''}
                                    ${caseItem.enabled_questions.taste_note_2 ? '<span style="background: #2ecc71; color: white; padding: 0.2rem 0.4rem; border-radius: 3px; font-size: 0.8rem;">Nota 2</span>' : ''}
                                    ${caseItem.enabled_questions.favorite_coffee ? '<span style="background: #2ecc71; color: white; padding: 0.2rem 0.4rem; border-radius: 3px; font-size: 0.8rem;">Favorito</span>' : ''}
                                    ${caseItem.enabled_questions.brewing_method ? '<span style="background: #2ecc71; color: white; padding: 0.2rem 0.4rem; border-radius: 3px; font-size: 0.8rem;">Método</span>' : ''}
                                </div>
                            </div>
                        ` : ''}
                        <small style="opacity: 0.6;">Creado: ${createdDate} | Cafés: ${coffeeCount}</small>
                    </div>
                    <div style="margin-left: 1rem;">
                        <button onclick="toggleCaseStatus('${caseItem.id}', ${!caseItem.is_active})" class="cta-button" 
                                style="background: ${caseItem.is_active ? '#e74c3c' : '#2ecc71'}; margin-bottom: 0.5rem; padding: 0.5rem 1rem; position: relative; z-index: 10; pointer-events: auto;">
                            ${caseItem.is_active ? 'Desactivar' : 'Activar'}
                        </button><br>
                        <button onclick="editCase('${caseItem.id}')" class="cta-button" 
                                style="background: #d4af37; margin-bottom: 0.5rem; padding: 0.5rem 1rem; position: relative; z-index: 10; pointer-events: auto;">
                            Editar
                        </button><br>
                        <button onclick="deleteCase('${caseItem.id}')" class="cta-button" 
                                style="background: #e74c3c; padding: 0.5rem 1rem; position: relative; z-index: 10; pointer-events: auto;">
                            Eliminar
                        </button>
                    </div>
                </div>
            </div>
        `;
    });
    
    html += '</div>';
    container.innerHTML = html;
}

async function toggleCaseStatus(caseId, newStatus) {
    try {
        await API.put(`${API_CONFIG.ENDPOINTS.ADMIN_CASES}/${caseId}`, {
            is_active: newStatus
        });
        showAdminNotification(`Caso ${newStatus ? 'activado' : 'desactivado'} exitosamente`, 'success');
        loadAllCases();
    } catch (error) {
        console.error('Error updating case status:', error);
        showAdminNotification('Error al actualizar el estado del caso', 'error');
    }
}

async function editCase(caseId) {
    try {
        // Fetch case data
        const response = await API.get(`${API_CONFIG.ENDPOINTS.ADMIN_CASES}/${caseId}`);
        const caseData = response.case;
        
        // Hide other forms
        document.getElementById('createCaseForm').classList.remove('admin-section__form--visible');
        
        // Show edit form first
        document.getElementById('editCaseForm').classList.add('admin-section__form--visible');
        
        // Wait a moment for the form to be rendered
        setTimeout(async () => {
            try {
                // Populate form with existing data
                populateEditForm(caseData);
                
                // Load dropdowns for edit form first
                await loadEditFormDropdowns();
                
                // Add a small delay to ensure dropdowns are fully rendered
                setTimeout(() => {
                    // Set dropdown values after dropdowns are populated
                    setEditFormDropdownValues(caseData);
                }, 200);
                
            } catch (error) {
                console.error('Error populating edit form:', error);
                showAdminNotification('Error al poblar el formulario de edición', 'error');
            }
        }, 50);
        
    } catch (error) {
        console.error('Error loading case for edit:', error);
        showAdminNotification('Error al cargar el caso para editar', 'error');
    }
}

function populateEditForm(caseData) {
    try {
        // Populating edit form with case data
        
        // Check if edit form is visible
        const editForm = document.getElementById('editCaseForm');
        if (!editForm || editForm.style.display === 'none') {
            console.error('Edit form is not visible or does not exist');
            return;
        }
        
        // Basic case info
        const editCaseId = document.getElementById('editCaseId');
        const editCaseName = document.getElementById('editCaseName');
        const editCaseDescription = document.getElementById('editCaseDescription');
        const editCaseIsActive = document.getElementById('editCaseIsActive');
    
    if (editCaseId) {
        editCaseId.value = caseData.id;
    } else {
        console.error('editCaseId element not found');
    }
    
    if (editCaseName) {
        editCaseName.value = caseData.name || '';
    } else {
        console.error('editCaseName element not found');
    }
    
    if (editCaseDescription) {
        editCaseDescription.value = caseData.description || '';
    } else {
        console.error('editCaseDescription element not found');
    }
    
    if (editCaseIsActive) {
        editCaseIsActive.checked = caseData.is_active || false;
    } else {
        console.error('editCaseIsActive element not found');
    }
    
    // Coffee details (names and notes, dropdowns will be set after loading)
    for (let i = 1; i <= 4; i++) {
        const coffee = caseData.coffees && caseData.coffees[i-1];
        const nameElement = document.getElementById(`editCoffee${i}Name`);
        const notesElement = document.getElementById(`editCoffee${i}Notes`);
        
        // Coffee elements validated
        
        if (coffee) {
            if (nameElement) {
                nameElement.value = coffee.name || '';
            } else {
                console.error(`editCoffee${i}Name element not found`);
            }
            
            if (notesElement) {
                notesElement.value = coffee.tasting_notes || '';
            } else {
                console.error(`editCoffee${i}Notes element not found`);
            }
        } else {
            // Clear fields if no coffee data
            if (nameElement) {
                nameElement.value = '';
            }
            if (notesElement) {
                notesElement.value = '';
            }
        }
    }
    
    // Question configuration
    const questions = caseData.enabled_questions || {
        region: true,
        variety: true,
        process: true,
        taste_note_1: true,
        taste_note_2: true,
        favorite_coffee: true,
        brewing_method: true
    };
    
    const questionElements = [
        { id: 'editQuestionRegion', value: questions.region },
        { id: 'editQuestionVariety', value: questions.variety },
        { id: 'editQuestionProcess', value: questions.process },
        { id: 'editQuestionTasteNote1', value: questions.taste_note_1 },
        { id: 'editQuestionTasteNote2', value: questions.taste_note_2 },
        { id: 'editQuestionFavoriteCoffee', value: questions.favorite_coffee },
        { id: 'editQuestionBrewingMethod', value: questions.brewing_method }
    ];
    
    questionElements.forEach(({ id, value }) => {
        const element = document.getElementById(id);
        if (element) {
            element.checked = value;
        } else {
            console.error(`${id} element not found`);
        }
    });
        
    } catch (error) {
        console.error('Error in populateEditForm:', error);
        throw error; // Re-throw to be caught by the calling function
    }
}

function setEditFormDropdownValues(caseData) {
    try {
        // Setting edit form dropdown values
        
        // Set coffee dropdown values after dropdowns are populated
        for (let i = 1; i <= 4; i++) {
        const coffee = caseData.coffees && caseData.coffees[i-1];
        if (coffee) {
            // Setting coffee data for form
            
            const regionSelect = document.getElementById(`editCoffee${i}Region`);
            const varietySelect = document.getElementById(`editCoffee${i}Variety`);
            const processSelect = document.getElementById(`editCoffee${i}Process`);
            
            // Processing coffee attributes
            
            if (regionSelect) {
                // Setting region dropdown
                
                // Try to set by value first
                regionSelect.value = coffee.region || '';
                
                // If value didn't match, try to find by label/text
                if (regionSelect.value === '' && coffee.region) {
                    const matchingOption = Array.from(regionSelect.options).find(option => 
                        option.textContent.toLowerCase() === coffee.region.toLowerCase()
                    );
                    if (matchingOption) {
                        regionSelect.value = matchingOption.value;
                        // Found region match
                    }
                }
                
                // Region set successfully
            }
            if (varietySelect) {
                // Setting variety dropdown
                
                // Try to set by value first
                varietySelect.value = coffee.variety || '';
                
                // If value didn't match, try to find by label/text
                if (varietySelect.value === '' && coffee.variety) {
                    const matchingOption = Array.from(varietySelect.options).find(option => 
                        option.textContent.toLowerCase() === coffee.variety.toLowerCase()
                    );
                    if (matchingOption) {
                        varietySelect.value = matchingOption.value;
                        // Found variety match
                    }
                }
                
                // Variety set successfully
            }
            if (processSelect) {
                // Setting process dropdown
                
                // Try to set by value first
                processSelect.value = coffee.process || '';
                
                // If value didn't match, try to find by label/text
                if (processSelect.value === '' && coffee.process) {
                    const matchingOption = Array.from(processSelect.options).find(option => 
                        option.textContent.toLowerCase() === coffee.process.toLowerCase()
                    );
                    if (matchingOption) {
                        processSelect.value = matchingOption.value;
                        // Found process match
                    }
                }
                
                // Process set successfully
            }
        }
    }
    
    } catch (error) {
        console.error('Error in setEditFormDropdownValues:', error);
        throw error; // Re-throw to be caught by the calling function
    }
}

async function loadEditFormDropdowns() {
    try {
        // Loading edit form dropdowns
        const data = await API.get(API_CONFIG.ENDPOINTS.CATALOG);
        const catalog = data.catalog;
        // Catalog data loaded for edit form
        
        // Populate dropdowns for each coffee in edit form
        for (let i = 1; i <= 4; i++) {
            // Populating edit form dropdowns for coffee
            populateCaseDropdown(`editCoffee${i}Region`, catalog.region || []);
            populateCaseDropdown(`editCoffee${i}Variety`, catalog.variety || []);
            populateCaseDropdown(`editCoffee${i}Process`, catalog.process || []);
            
            // Verify dropdowns were populated
            const regionSelect = document.getElementById(`editCoffee${i}Region`);
            const varietySelect = document.getElementById(`editCoffee${i}Variety`);
            const processSelect = document.getElementById(`editCoffee${i}Process`);
            
            // Coffee dropdowns populated successfully
        }
        
        // Edit form dropdowns loading complete
        
    } catch (error) {
        console.error('Failed to load catalog data for edit form:', error);
    }
}

function cancelEditCase() {
    document.getElementById('editCaseForm').classList.remove('admin-section__form--visible');
}

async function updateCase() {
    const caseId = document.getElementById('editCaseId').value;
    const name = document.getElementById('editCaseName').value.trim();
    const description = document.getElementById('editCaseDescription').value.trim();
    const isActive = document.getElementById('editCaseIsActive').checked;
    
    // Get the original case data to preserve coffee IDs
    let originalCoffeeIds = [];
    try {
        const response = await API.get(`${API_CONFIG.ENDPOINTS.ADMIN_CASES}/${caseId}`);
        const originalCase = response.case;
        originalCoffeeIds = originalCase.coffees ? originalCase.coffees.map(c => c.id) : [];
        // Processing coffee IDs for update
    } catch (error) {
        console.error('Error getting original case data:', error);
        // If we can't get original IDs, we'll use generated ones as fallback
    }
    
    if (!name || !description) {
        showAdminNotification('Nombre y descripción son requeridos', 'error');
        return;
    }
    
    // Collect coffee data
    const coffees = [];
    for (let i = 1; i <= 4; i++) {
        const coffeeName = document.getElementById(`editCoffee${i}Name`).value.trim();
        const region = document.getElementById(`editCoffee${i}Region`).value;
        const variety = document.getElementById(`editCoffee${i}Variety`).value;
        const process = document.getElementById(`editCoffee${i}Process`).value;
        const notes = document.getElementById(`editCoffee${i}Notes`).value.trim();
        
        if (!coffeeName || !region || !variety || !process || !notes) {
            showAdminNotification(`Todos los campos del Café #${i} son requeridos`, 'error');
            return;
        }
        
        coffees.push({
            id: originalCoffeeIds[i-1] || `coffee_${i}`, // Use original ID if available, fallback to generated
            name: coffeeName,
            region: region,
            variety: variety,
            process: process,
            tasting_notes: notes
        });
    }
    
    // Collect enabled questions
    const enabledQuestions = {
        region: document.getElementById('editQuestionRegion').checked,
        variety: document.getElementById('editQuestionVariety').checked,
        process: document.getElementById('editQuestionProcess').checked,
        taste_note_1: document.getElementById('editQuestionTasteNote1').checked,
        taste_note_2: document.getElementById('editQuestionTasteNote2').checked,
        favorite_coffee: document.getElementById('editQuestionFavoriteCoffee').checked,
        brewing_method: document.getElementById('editQuestionBrewingMethod').checked
    };

    const updatedCase = {
        name: name,
        description: description,
        is_active: isActive,
        coffees: coffees,
        enabled_questions: enabledQuestions
    };
    
    try {
        await API.put(`${API_CONFIG.ENDPOINTS.ADMIN_CASES}/${caseId}`, updatedCase);
        showAdminNotification('Caso actualizado exitosamente', 'success');
        cancelEditCase();
        loadAllCases();
    } catch (error) {
        console.error('Error updating case:', error);
        showAdminNotification('Error al actualizar el caso', 'error');
    }
}

async function deleteCase(caseId) {
    if (confirm('¿Estás seguro de que quieres eliminar este caso? Esta acción no se puede deshacer.')) {
        try {
            await API.delete(`${API_CONFIG.ENDPOINTS.ADMIN_CASES}/${caseId}`);
            showAdminNotification('Caso eliminado exitosamente', 'success');
            loadAllCases();
        } catch (error) {
            console.error('Error deleting case:', error);
            showAdminNotification('Error al eliminar el caso', 'error');
        }
    }
}

// Order Management Functions

function showCreateOrderForm() {
    document.getElementById('createOrderForm').classList.add('admin-section__form--visible');
    clearOrderForm();
}

function cancelCreateOrder() {
    document.getElementById('createOrderForm').classList.remove('admin-section__form--visible');
}

function clearOrderForm() {
    document.getElementById('orderUserSelect').value = '';
    document.getElementById('orderCaseSelect').value = '';
    document.getElementById('orderContactInfo').value = '';
    document.getElementById('orderAmount').value = '';
    document.getElementById('orderStatus').value = 'pending';
}

async function loadOrderFormDropdowns() {
    try {
        // Load users
        const usersData = await API.get(API_CONFIG.ENDPOINTS.ADMIN_USERS);
        populateOrderDropdown('orderUserSelect', usersData.users || [], 'name', 'id');
        
        // Load cases
        const casesData = await API.get(API_CONFIG.ENDPOINTS.ADMIN_CASES);
        populateOrderDropdown('orderCaseSelect', casesData.cases || [], 'name', 'id');
        
    } catch (error) {
        console.error('Failed to load order form dropdowns:', error);
    }
}

function populateOrderDropdown(selectId, items, labelField, valueField) {
    const select = document.getElementById(selectId);
    if (!select) return;
    
    // Keep the first "Select" option
    const firstOption = select.firstElementChild;
    select.innerHTML = '';
    select.appendChild(firstOption);
    
    // Add items
    items.forEach(item => {
        const option = document.createElement('option');
        option.value = item[valueField];
        option.textContent = item[labelField];
        select.appendChild(option);
    });
}

async function createOrder() {
    const userId = document.getElementById('orderUserSelect').value;
    const caseId = document.getElementById('orderCaseSelect').value;
    const contactInfo = document.getElementById('orderContactInfo').value.trim();
    const amount = parseInt(document.getElementById('orderAmount').value) || 0;
    const status = document.getElementById('orderStatus').value;
    
    if (!userId || !caseId) {
        showAdminNotification('Usuario y caso son requeridos', 'error');
        return;
    }
    
    const newOrder = {
        user_id: userId,
        case_id: caseId,
        contact_info: contactInfo,
        total_amount: amount,
        status: status
    };
    
    try {
        const response = await API.post(API_CONFIG.ENDPOINTS.ADMIN_ORDERS, newOrder);
        showAdminNotification(`Orden creada exitosamente. ID de orden: ${response.customer_order_id}`, 'success');
        cancelCreateOrder();
        loadAllOrders();
    } catch (error) {
        console.error('Error creating order:', error);
        showAdminNotification('Error al crear la orden', 'error');
    }
}

async function loadAllOrders() {
    const container = document.getElementById('ordersList');
    
    // Show loading
    container.innerHTML = '<p>Cargando órdenes...</p>';
    
    try {
        const data = await API.get(`${API_CONFIG.ENDPOINTS.ADMIN_ORDERS}?limit=50`);
        displayOrders(data.orders || []);
    } catch (error) {
        console.error('Error loading orders:', error);
        showAdminNotification('Error al cargar las órdenes', 'error');
        container.innerHTML = '<p>Error al cargar las órdenes</p>';
    }
}

function displayOrders(orders) {
    const container = document.getElementById('ordersList');
    
    if (orders.length === 0) {
        container.innerHTML = '<p>No hay órdenes creadas</p>';
        return;
    }
    
    let html = '<div style="display: grid; gap: 1rem;">';
    
    orders.forEach(order => {
        const createdDate = new Date(order.created_at).toLocaleDateString('es-ES');
        const statusColor = getStatusColor(order.status);
        const usedBadge = order.is_submission_used ? 
            '<span style="background: #e74c3c; color: white; padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.8rem;">USADO</span>' :
            '<span style="background: #2ecc71; color: white; padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.8rem;">DISPONIBLE</span>';
        
        html += `
            <div style="background: rgba(0,0,0,0.3); padding: 1rem; border-radius: 8px; position: relative; z-index: 5;">
                <div style="display: flex; justify-content: space-between; align-items: start;">
                    <div style="flex: 1;">
                        <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem;">
                            <strong style="color: #d4af37;">ID: ${order.order_id}</strong>
                            <span style="background: ${statusColor}; color: white; padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.8rem;">${order.status.toUpperCase()}</span>
                            ${usedBadge}
                        </div>
                        <div style="margin-bottom: 0.5rem;">
                            <strong>Usuario:</strong> ${order.user_name || 'Sin nombre'} (${order.user_id})<br>
                            <strong>Caso:</strong> ${order.case_name || 'Sin nombre'}<br>
                            <strong>Monto:</strong> ₡${order.total_amount?.toLocaleString() || '0'}
                        </div>
                        ${order.contact_info ? `
                            <div style="background: rgba(212, 175, 55, 0.1); padding: 0.5rem; border-radius: 4px; margin-top: 0.5rem;">
                                <small><strong>Contacto:</strong> ${order.contact_info}</small>
                            </div>
                        ` : ''}
                        ${order.is_submission_used && order.submission_used_at ? `
                            <div style="margin-top: 0.5rem; opacity: 0.8; font-size: 0.9rem;">
                                <small>Usado el: ${new Date(order.submission_used_at).toLocaleDateString('es-ES')}</small>
                            </div>
                        ` : ''}
                        <small style="opacity: 0.6;">Creado: ${createdDate}</small>
                    </div>
                    <div style="margin-left: 1rem;">
                        <select onchange="updateOrderStatus('${order.id}', this.value)" style="margin-bottom: 0.5rem; position: relative; z-index: 10; pointer-events: auto;">
                            <option value="pending" ${order.status === 'pending' ? 'selected' : ''}>Pendiente</option>
                            <option value="confirmed" ${order.status === 'confirmed' ? 'selected' : ''}>Confirmada</option>
                            <option value="shipped" ${order.status === 'shipped' ? 'selected' : ''}>Enviada</option>
                            <option value="delivered" ${order.status === 'delivered' ? 'selected' : ''}>Entregada</option>
                        </select><br>
                        <button onclick="copyOrderId('${order.order_id}')" class="cta-button" 
                                style="background: #3498db; padding: 0.5rem 1rem; position: relative; z-index: 10; pointer-events: auto;">
                            Copiar ID
                        </button>
                    </div>
                </div>
            </div>
        `;
    });
    
    html += '</div>';
    container.innerHTML = html;
}

function getStatusColor(status) {
    switch (status) {
        case 'pending': return '#f39c12';
        case 'confirmed': return '#3498db';
        case 'shipped': return '#9b59b6';
        case 'delivered': return '#2ecc71';
        default: return '#666';
    }
}

async function updateOrderStatus(orderId, newStatus) {
    try {
        await API.put(`${API_CONFIG.ENDPOINTS.ADMIN_ORDERS}/${orderId}/status`, {
            status: newStatus
        });
        showAdminNotification(`Estado actualizado a: ${newStatus}`, 'success');
        loadAllOrders();
    } catch (error) {
        console.error('Error updating order status:', error);
        showAdminNotification('Error al actualizar el estado', 'error');
        loadAllOrders(); // Reload to reset the dropdown
    }
}

function copyOrderId(orderId) {
    navigator.clipboard.writeText(orderId).then(() => {
        showAdminNotification(`ID copiado: ${orderId}`, 'success');
    }).catch(() => {
        // Fallback for older browsers
        const textArea = document.createElement('textarea');
        textArea.value = orderId;
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);
        showAdminNotification(`ID copiado: ${orderId}`, 'success');
    });
}

// Add some interactive animations
document.addEventListener('DOMContentLoaded', function() {
    // Initialize router
    Router.init();
    
    // Handle navigation link clicks
    setupNavigationLinks();
    
    // Load catalog data on page load
    loadCatalogData();
    
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