# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Brew Detective** is a web-based coffee tasting game focused on Costa Rican coffee. It's a single-page application (SPA) that allows users to:
- Order mystery coffee boxes (4 Costa Rican coffees)
- Submit their coffee tasting analysis and guesses
- View leaderboards and track their progress as "detectives"
- Manage their profile and view statistics

## Architecture

This is a static website built with:
- **Frontend**: HTML, SASS-compiled CSS, and JavaScript (no frameworks)
- **Build Process**: SASS compilation with npm scripts
- **Deployment**: Docker containerized with Nginx
- **Structure**: Single-page application with client-side routing

### Key Files
- `src/index.html` - Main HTML file
- `src/sass/` - SASS source files organized in partials
- `src/css/main.css` - Compiled CSS output
- `src/js/main.js` - JavaScript functionality
- `src/images/` - Static images (logos, etc.)
- `package.json` - Build scripts and dependencies
- `docker-compose.yml` - Container setup with Nginx and development tools
- `nginx.conf` - Nginx configuration for serving the static site

## Development Setup

### Prerequisites
- Docker and Docker Compose
- Node.js and npm (if developing outside Docker)

### Running the Application
```bash
# Start the development environment (includes SASS watch)
docker-compose up -d

# Access the application
# http://localhost:8080
```

### SASS Development
```bash
# Build CSS once
npm run build:css

# Watch SASS files for changes (automatic compilation)
npm run watch:css

# Run full build process
npm run build
```

### Development Container Access
```bash
# Access the development container for any Node.js tools
docker-compose exec web-dev sh
```

### Stopping the Environment
```bash
docker-compose down
```

## Application Structure

The application uses JavaScript-based client-side routing with these main sections:
- **Home**: Landing page with game explanation
- **Order**: WhatsApp integration for ordering coffee boxes
- **Submit**: Form for submitting coffee tasting analysis
- **Leaderboard**: Ranking system for players
- **Profile**: User statistics and badges

### Key Components
- Clean separation of HTML, CSS, and JavaScript
- Mobile-responsive design with hamburger menu
- Form validation and mock data submission
- Simulated leaderboard updates and user statistics
- Local storage for user data (profile information)
- Intersection Observer for scroll animations

## Business Logic

The application simulates a coffee detective game where:
1. Users order mystery coffee boxes via WhatsApp
2. They taste and analyze 4 different Costa Rican coffees
3. Submit their guesses about origin, variety, process, and roast level
4. Get scored and ranked against other "detectives"
5. Earn badges and track their progress

## Data Flow

Currently all data is mock/simulated:
- Form submissions show success/error messages but don't persist
- Leaderboard data is hardcoded
- Profile statistics are simulated
- No backend API integration

## Notable Features

- **WhatsApp Integration**: Direct link to order coffee boxes
- **Costa Rican Coffee Focus**: Specific regions, varieties, and processes
- **Gamification**: Points, badges, rankings, and detective levels
- **Responsive Design**: Works on mobile and desktop
- **Progressive Enhancement**: Graceful fallbacks for JavaScript disabled

## Technical Notes

- **SASS Architecture**: Organized with partials for maintainability
  - `base/` - Variables, mixins, reset styles
  - `layout/` - Header, footer, hero, sections
  - `components/` - Cards, forms, leaderboard
  - `utilities/` - Animations and helper classes
- **Build Process**: SASS compilation with npm scripts
- **Development**: Docker container with automatic SASS watching
- **CSS Features**: Grid, Flexbox, backdrop-filter, modern CSS
- **JavaScript**: Vanilla JS with modern ES6+ features
- **Deployment**: Nginx serves static files with optimized caching