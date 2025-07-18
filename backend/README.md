# Brew Detective Backend

A Golang REST API backend for the Brew Detective coffee tasting game, built with Gin framework and Google Cloud Firestore.

## Features

- **Coffee Case Management**: Create and manage coffee mystery cases
- **Submission Processing**: Handle user submissions with scoring and accuracy calculation
- **Leaderboard System**: Real-time rankings with points and badges
- **User Profiles**: User registration and profile management
- **Order Management**: Coffee case ordering system
- **CORS Support**: Configured for GitHub Pages frontend

## Technology Stack

- **Go 1.21+** - Programming language
- **Gin** - HTTP web framework
- **Google Cloud Firestore** - NoSQL database
- **Google Cloud Run** - Serverless deployment
- **Docker** - Containerization

## API Endpoints

### Cases
- `GET /api/v1/cases` - Get all active coffee cases
- `GET /api/v1/cases/:id` - Get specific case details

### Submissions
- `POST /api/v1/submissions` - Submit a case solution

### Leaderboard
- `GET /api/v1/leaderboard` - Get current leaderboard

### Users
- `GET /api/v1/users/:id` - Get user profile
- `PUT /api/v1/users/:id` - Update user profile

### Orders
- `POST /api/v1/orders` - Create new order
- `GET /api/v1/orders/:id` - Get order details
- `PUT /api/v1/orders/:id/status` - Update order status

## Local Development

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Set environment variables**:
   ```bash
   export GOOGLE_CLOUD_PROJECT="your-project-id"
   export GOOGLE_APPLICATION_CREDENTIALS="path/to/service-account.json"
   export PORT=8080
   ```

3. **Run the server**:
   ```bash
   go run cmd/server/main.go
   ```

## Deployment

Deploy to Google Cloud Run:

```bash
gcloud run deploy brew-detective-backend \
  --source . \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars GOOGLE_CLOUD_PROJECT=your-project-id
```

## Data Models

- **User**: Detective profiles with stats and badges
- **CoffeeCase**: Mystery coffee cases with multiple coffees
- **Submission**: User answers and scoring
- **Order**: Coffee case orders
- **LeaderboardEntry**: Ranking information

## Scoring System

The scoring system evaluates submissions based on:
- **Accuracy**: Percentage of correct answers
- **Points**: Base points multiplied by accuracy and coffee count
- **Badges**: Achievement-based rewards

## CORS Configuration

The API is configured to accept requests from:
- GitHub Pages domains
- Local development servers (localhost:3000, localhost:8080)

Update the CORS configuration in `cmd/server/main.go` to match your frontend domain.

## Environment Variables

- `GOOGLE_CLOUD_PROJECT`: GCP project ID
- `GOOGLE_APPLICATION_CREDENTIALS`: Path to service account JSON (local only)
- `PORT`: Server port (default: 8080)