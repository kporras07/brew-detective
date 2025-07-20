# 🔍 Brew Detective

A web-based coffee tasting game focused on Costa Rican specialty coffee. Test your palate by analyzing mystery coffee samples and compete with other coffee detectives!

[![Live Demo](https://img.shields.io/badge/Live%20Demo-GitHub%20Pages-green)](https://kporras07.github.io/brew-detective)
[![Backend API](https://img.shields.io/badge/Backend%20API-Google%20Cloud%20Run-blue)](https://brew-detective-backend-1087966598090.us-central1.run.app)

## 🎯 Features

- **Mystery Coffee Boxes**: Order curated sets of 4 Costa Rican specialty coffees
- **Tasting Analysis**: Submit detailed tasting notes and origin guesses
- **Authentication**: Secure Google OAuth login system
- **Leaderboard**: Compete with other coffee detectives
- **Progressive Detective Levels**: Advance from Rookie to Master Detective
- **Real-time Scoring**: Get instant feedback on your coffee knowledge
- **Mobile Responsive**: Works seamlessly on all devices

## 🏗️ Architecture

### Frontend
- **Framework**: Vanilla JavaScript (no dependencies)
- **Styling**: SASS with modern CSS features
- **Build System**: npm scripts for SASS compilation
- **Deployment**: GitHub Pages with automated CI/CD
- **Authentication**: Google OAuth 2.0 integration

### Backend
- **Language**: Go 1.21
- **Framework**: Gin (HTTP web framework)
- **Database**: Google Cloud Firestore (NoSQL)
- **Authentication**: JWT tokens with Google OAuth
- **Deployment**: Google Cloud Run (serverless containers)
- **Security**: CSRF protection, secure cookies, environment-aware configuration

## 🚀 Live Demo

- **Frontend**: [https://kporras07.github.io/brew-detective](https://kporras07.github.io/brew-detective)
- **Backend API**: [https://brew-detective-backend-1087966598090.us-central1.run.app](https://brew-detective-backend-1087966598090.us-central1.run.app)

## 🛠️ Development Setup

### Prerequisites

- **Node.js** 16+ (for SASS compilation)
- **Go** 1.21+ (for backend development)
- **Docker** (for containerized development)
- **Google Cloud SDK** (for deployment)

### Frontend Development

1. **Clone the repository**:
   ```bash
   git clone https://github.com/kporras07/brew-detective.git
   cd brew-detective
   ```

2. **Install dependencies**:
   ```bash
   npm install
   ```

3. **Start development with SASS watching**:
   ```bash
   # Using Docker (recommended)
   docker-compose up -d
   
   # Or using npm directly
   npm run watch:css
   ```

4. **Access the application**:
   - Frontend: [http://localhost:8080](http://localhost:8080)
   - Development container: `docker-compose exec web-dev sh`

### Backend Development

1. **Navigate to backend directory**:
   ```bash
   cd backend
   ```

2. **Set up environment variables**:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Install Go dependencies**:
   ```bash
   go mod download
   ```

4. **Run the backend server**:
   ```bash
   go run cmd/server/main.go
   ```

5. **Backend will be available at**: [http://localhost:8888](http://localhost:8888)

## 🔧 Environment Configuration

### Frontend Configuration

Update `src/js/config.js` to point to your backend:

```javascript
const API_CONFIG = {
    // For production
    BASE_URL: 'https://your-backend-url.run.app',
    
    // For local development
    // BASE_URL: 'http://localhost:8888',
};
```

### Backend Environment Variables

Create `backend/.env` with:

```env
# Google Cloud
GOOGLE_CLOUD_PROJECT=your-project-id
FIRESTORE_DATABASE_ID=your-database-name

# OAuth Configuration
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=https://your-backend-url/auth/google/callback

# JWT Security
JWT_SECRET=your-super-secret-jwt-key

# Application Configuration
FRONTEND_URL=https://your-frontend-url
GIN_MODE=release
```

## 🔐 Google OAuth Setup

1. **Create Google Cloud Project**:
   - Go to [Google Cloud Console](https://console.cloud.google.com)
   - Create a new project or select existing one

2. **Enable APIs**:
   - Enable Google+ API
   - Enable Google OAuth2 API

3. **Create OAuth 2.0 Credentials**:
   - Go to APIs & Services → Credentials
   - Create OAuth 2.0 Client ID
   - Add authorized origins and redirect URIs

4. **Configure OAuth Consent Screen**:
   - Set application name and description
   - Add authorized domains

### Required OAuth Configuration

**Authorized JavaScript Origins**:
```
https://your-domain.github.io
https://your-backend-url.run.app
http://localhost:8080
```

**Authorized Redirect URIs**:
```
https://your-backend-url.run.app/auth/google/callback
http://localhost:8888/auth/google/callback
```

## 🚀 Deployment

### Frontend Deployment (GitHub Pages)

The frontend automatically deploys to GitHub Pages via GitHub Actions:

1. **Enable GitHub Pages** in repository settings
2. **Push to main branch** - automatic deployment triggers
3. **SASS compilation** happens automatically in CI/CD
4. **Live site** available at `https://username.github.io/repository-name`

### Backend Deployment (Google Cloud Run)

1. **Build and push Docker image**:
   ```bash
   cd backend
   
   # Build for AMD64 architecture
   docker build --platform linux/amd64 -t gcr.io/your-project/brew-detective-backend .
   
   # Push to registry
   docker push gcr.io/your-project/brew-detective-backend
   ```

2. **Deploy to Cloud Run**:
   ```bash
   gcloud run deploy brew-detective-backend \
     --image gcr.io/your-project/brew-detective-backend \
     --region us-central1 \
     --allow-unauthenticated \
     --set-env-vars GOOGLE_CLOUD_PROJECT=your-project
   ```

3. **Configure secrets** in Google Secret Manager:
   ```bash
   echo "your-google-client-secret" | gcloud secrets create google-client-secret --data-file=-
   echo "your-jwt-secret" | gcloud secrets create jwt-secret --data-file=-
   ```

## 📁 Project Structure

```
brew-detective/
├── src/                          # Frontend source code
│   ├── index.html               # Main HTML file
│   ├── css/main.css             # Compiled CSS (generated)
│   ├── sass/                    # SASS source files
│   │   ├── main.scss           # Main SASS entry point
│   │   ├── base/               # Variables, mixins, reset
│   │   ├── layout/             # Header, footer, sections
│   │   ├── components/         # Cards, forms, leaderboard
│   │   └── utilities/          # Animations, helpers
│   ├── js/                     # JavaScript files
│   │   ├── main.js             # Main application logic
│   │   ├── config.js           # API configuration
│   │   └── auth.js             # Authentication handling
│   └── images/                 # Static assets
├── backend/                     # Backend source code
│   ├── cmd/server/main.go      # Server entry point
│   ├── internal/               # Internal packages
│   │   ├── auth/               # Authentication logic
│   │   ├── database/           # Database connections
│   │   ├── handlers/           # HTTP handlers
│   │   └── models/             # Data models
│   ├── Dockerfile              # Container configuration
│   └── deploy.yaml             # Cloud Run deployment config
├── docker-compose.yml          # Development environment
├── nginx.conf                  # Nginx configuration
├── package.json                # Node.js dependencies
└── README.md                   # This file
```

## 🧪 Development Commands

### Frontend Commands

```bash
# Install dependencies
npm install

# Build CSS once
npm run build:css

# Watch SASS files for changes
npm run watch:css

# Start development environment
docker-compose up -d

# Stop development environment
docker-compose down
```

### Backend Commands

```bash
# Run development server
go run cmd/server/main.go

# Build binary
go build -o main cmd/server/main.go

# Run tests
go test ./...

# Build Docker image
docker build -t brew-detective-backend .
```

## 🎯 Scoring System

The game uses a dynamic scoring system based on enabled questions for each case:

### Base Scoring Formula
```
Final Score = (Base Points × Accuracy × Number of Coffees) + Bonus Points
```

### Points Breakdown
- **Base Points**: 100 per perfect coffee analysis
- **Coffee Questions**: Each correct answer contributes to accuracy
  - Region identification
  - Coffee variety
  - Processing method  
  - Tasting note 1
  - Tasting note 2

### Example Scoring (4 coffees, all questions enabled)
- **Total Questions**: 20 (5 questions × 4 coffees)
- **Points per Coffee Question**: 20 points (100 ÷ 5 questions)
- **Perfect Coffee Score**: 100 points (20 × 5 correct answers)
- **Maximum Base Score**: 400 points (100 × 4 coffees)

### Bonus Points
- **Favorite Coffee**: +50 points
- **Brewing Method**: +50 points
- **Maximum Total Score**: 500 points

### Accuracy Calculation
```
Accuracy = Correct Answers ÷ Total Enabled Questions
```

The scoring adapts automatically when admins enable/disable questions for different cases, ensuring fair competition regardless of case complexity.

## 🔒 Security Features

- **Google OAuth 2.0** authentication
- **JWT tokens** with secure signing
- **CSRF protection** via OAuth state parameter
- **Secure cookies** with HttpOnly and Secure flags
- **Environment-aware security** (HTTPS in production, HTTP in development)
- **Input validation** and sanitization
- **Rate limiting** on API endpoints

## 🌟 Coffee Education

Learn about Costa Rican coffee regions featured in the game:

- **Tarrazú**: High altitude, bright acidity, full body
- **Central Valley**: Balanced, chocolatey, medium body
- **West Valley**: Wine-like acidity, fruity notes
- **Brunca**: Lower altitude, nutty, smooth
- **Turrialba**: Volcanic soil, earthy, robust
- **Orosi**: Floral, delicate, tea-like
- **Tres Ríos**: Complex, wine-like, exceptional quality
- **Guanacaste**: Dry processed, fruity, unique terroir

## 🤝 Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Commit your changes**: `git commit -m 'Add amazing feature'`
4. **Push to the branch**: `git push origin feature/amazing-feature`
5. **Open a Pull Request**

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Costa Rican coffee producers for inspiring this project
- Google Cloud Platform for hosting infrastructure
- The coffee community for sharing knowledge about specialty coffee
- Open source contributors who made this project possible

## 📞 Contact

**Developer**: Kevin Porras  
**GitHub**: [@kporras07](https://github.com/kporras07)  
**Project**: [Brew Detective](https://github.com/kporras07/brew-detective)

---

*Made with ☕ and ❤️ in Costa Rica*