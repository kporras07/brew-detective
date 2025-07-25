# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Brew Detective** is a web-based coffee tasting game focused on Costa Rican coffee. It's a single-page application (SPA) that allows users to:
- Order mystery coffee boxes (4 Costa Rican coffees)
- Submit their coffee tasting analysis and guesses
- View leaderboards and track their progress as "detectives"
- Manage their profile and view statistics

## Architecture

This is a full-stack application with:
- **Frontend**: HTML, SASS-compiled CSS, and JavaScript (no frameworks)
- **Backend**: Golang REST API with Gin framework
- **Database**: Google Cloud Firestore (NoSQL)
- **Build Process**: SASS compilation with npm scripts
- **Deployment**: 
  - Frontend: GitHub Pages
  - Backend: Google Cloud Run
- **Structure**: Single-page application with client-side routing

### Key Files

**Frontend:**
- `src/index.html` - Main HTML file
- `src/sass/` - SASS source files organized in partials
- `src/css/main.css` - Compiled CSS output
- `src/js/main.js` - JavaScript functionality
- `src/js/config.js` - API configuration and helper functions
- `src/images/` - Static images (logos, etc.)
- `package.json` - Build scripts and dependencies
- `docker-compose.yml` - Container setup with Nginx and development tools
- `nginx.conf` - Nginx configuration for serving the static site

**Backend:**
- `backend/cmd/server/main.go` - Main server entry point
- `backend/internal/models/models.go` - Data models
- `backend/internal/handlers/` - API endpoint handlers
- `backend/internal/database/firestore.go` - Firestore client setup
- `backend/Dockerfile` - Container configuration for Cloud Run
- `backend/deploy.yaml` - Cloud Run deployment configuration

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

## Deployment

### GitHub Pages via GitHub Actions

The project is set up for automatic deployment to GitHub Pages using GitHub Actions:

1. **Push to main branch** - Triggers automatic deployment
2. **Build process** - Compiles SASS to CSS automatically
3. **Deploy to GitHub Pages** - Serves the `src/` directory

### Manual Deployment Steps

1. **Enable GitHub Pages** in your repository settings:
   - Go to Settings → Pages
   - Source: GitHub Actions

2. **Push your code** to the main branch:
   ```bash
   git add .
   git commit -m "Initial deployment"
   git push origin main
   ```

3. **GitHub Actions will**:
   - Install Node.js dependencies
   - Compile SASS to CSS
   - Deploy to GitHub Pages
   - Provide a live URL

### Build for Production
```bash
# Build compressed CSS for production
npm run build:css

# Build with source maps for debugging
npm run build:css:dev
```

## Backend Setup

### Prerequisites
- Go 1.21 or higher
- Google Cloud SDK
- Docker (optional, for local development)

### Local Development
```bash
# Navigate to backend directory
cd backend

# Install dependencies
go mod download

# Set environment variables
export GOOGLE_CLOUD_PROJECT="your-project-id"
export GOOGLE_APPLICATION_CREDENTIALS="path/to/service-account.json"

# Run the server
go run cmd/server/main.go
```

### Cloud Run Deployment

1. **Setup Google Cloud Project**:
   ```bash
   gcloud config set project YOUR_PROJECT_ID
   gcloud services enable run.googleapis.com
   gcloud services enable firestore.googleapis.com
   ```

2. **Build and Deploy**:
   ```bash
   cd backend
   gcloud run deploy brew-detective-backend \
     --source . \
     --region us-central1 \
     --allow-unauthenticated \
     --set-env-vars GOOGLE_CLOUD_PROJECT=YOUR_PROJECT_ID
   ```

3. **Update Frontend Config**:
   - Update `src/js/config.js` with your Cloud Run URL
   - Update CORS configuration in `backend/cmd/server/main.go`

### API Endpoints

- `GET /api/v1/cases` - Get all coffee cases
- `GET /api/v1/cases/:id` - Get specific case
- `POST /api/v1/submissions` - Submit case solution
- `GET /api/v1/leaderboard` - Get leaderboard
- `GET /api/v1/users/:id` - Get user profile
- `PUT /api/v1/users/:id` - Update user profile
- `POST /api/v1/orders` - Create order
- `GET /api/v1/orders/:id` - Get order details

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