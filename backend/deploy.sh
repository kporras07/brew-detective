#!/bin/bash

# Brew Detective Backend Deployment Script
set -e

# Configuration
PROJECT_ID="brew-detective"
REGION="us-central1"
SERVICE_NAME="brew-detective-backend"
IMAGE_NAME="us-central1-docker.pkg.dev/${PROJECT_ID}/brew-detective-repo/${SERVICE_NAME}"

echo "🚀 Starting deployment for Brew Detective Backend..."

# Step 1: Build Docker image for linux/amd64 platform
echo "📦 Building Docker image for linux/amd64..."
docker build --platform linux/amd64 -t ${IMAGE_NAME}:latest .

# Step 2: Push to Artifact Registry
echo "⬆️  Pushing image to Artifact Registry..."
docker push ${IMAGE_NAME}:latest

# Step 3: Deploy to Cloud Run using manifest
echo "🌐 Deploying to Cloud Run..."
gcloud run services replace deploy.yaml --region=${REGION}

# Step 4: Get the service URL
echo "✅ Deployment complete!"
SERVICE_URL=$(gcloud run services describe ${SERVICE_NAME} --region=${REGION} --format='value(status.url)')
echo "🔗 Service URL: ${SERVICE_URL}"

echo "🎉 Backend deployment successful!"