# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Brew Detective** is a web-based coffee tasting game focused on Costa Rican coffee. It's a single-page application (SPA) that allows users to:
- Order mystery coffee boxes (4 Costa Rican coffees)
- Submit their coffee tasting analysis and guesses
- View leaderboards and track their progress as "detectives"
- Manage their profile and view statistics

## Architecture

This is a simple static website built with:
- **Frontend**: Pure HTML, CSS, and JavaScript (no frameworks)
- **Deployment**: Docker containerized with Nginx
- **Structure**: Single-page application with client-side routing

### Key Files
- `src/index.html` - Complete application (HTML, CSS, JavaScript all in one file)
- `docker-compose.yml` - Container setup with Nginx and development tools
- `nginx.conf` - Nginx configuration for serving the static site

## Development Setup

### Running the Application
```bash
# Start the development environment
docker-compose up -d

# Access the application
# http://localhost:8080
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
- Single HTML file architecture with inline CSS and JavaScript
- Mobile-responsive design with hamburger menu
- Form validation and mock data submission
- Simulated leaderboard updates and user statistics
- Local storage for user data (profile information)

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

- No build process required - it's a static HTML file
- No package.json or npm dependencies
- Uses modern CSS features (Grid, Flexbox, backdrop-filter)
- Vanilla JavaScript with modern ES6+ features
- Docker setup provides consistent development environment