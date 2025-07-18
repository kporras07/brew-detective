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
}

// Submit form functionality
document.getElementById('submitForm').addEventListener('submit', function(e) {
    e.preventDefault();
    
    // Hide previous messages
    document.getElementById('submitSuccess').style.display = 'none';
    document.getElementById('submitError').style.display = 'none';
    
    // Collect form data
    const formData = {
        detectiveName: document.getElementById('detectiveName').value,
        caseId: document.getElementById('caseId').value,
        coffee1: {
            region: document.getElementById('coffee1_region').value,
            variety: document.getElementById('coffee1_variety').value,
            process: document.getElementById('coffee1_process').value,
            roast: document.getElementById('coffee1_roast').value,
            notes: document.getElementById('coffee1_notes').value
        },
        coffee2: {
            region: document.getElementById('coffee2_region').value,
            variety: document.getElementById('coffee2_variety').value,
            process: document.getElementById('coffee2_process').value,
            roast: document.getElementById('coffee2_roast').value,
            notes: document.getElementById('coffee2_notes').value
        },
        coffee3: {
            region: document.getElementById('coffee3_region').value,
            variety: document.getElementById('coffee3_variety').value,
            process: document.getElementById('coffee3_process').value,
            roast: document.getElementById('coffee3_roast').value,
            notes: document.getElementById('coffee3_notes').value
        },
        coffee4: {
            region: document.getElementById('coffee4_region').value,
            variety: document.getElementById('coffee4_variety').value,
            process: document.getElementById('coffee4_process').value,
            roast: document.getElementById('coffee4_roast').value,
            notes: document.getElementById('coffee4_notes').value
        },
        favoriteCoffe: document.getElementById('favorite_coffee').value,
        brewingMethod: document.getElementById('brewing_method').value
    };
    
    // Simulate submission processing
    setTimeout(() => {
        const random = Math.random();
        if (random > 0.1) { // 90% success rate
            document.getElementById('submitSuccess').style.display = 'block';
            document.getElementById('submitForm').reset();
            
            // Simulate adding to leaderboard
            setTimeout(() => {
                showPage('leaderboard');
            }, 2000);
        } else {
            document.getElementById('submitError').style.display = 'block';
        }
    }, 1000);
});

// Profile functionality
function saveProfile() {
    const name = document.getElementById('profileName').value;
    const email = document.getElementById('profileEmail').value;
    const level = document.getElementById('profileLevel').value;
    
    if (name && email) {
        alert('¡Perfil actualizado exitosamente, Detective ' + name + '!');
    } else {
        alert('Por favor completa todos los campos obligatorios.');
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