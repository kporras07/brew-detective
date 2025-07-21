#!/bin/bash

# Quick Deploy Script - Forces a new Cloud Run revision without rebuilding
# Useful when you want to restart the service or force a new deployment

set -e

PROJECT_ID="brew-detective"
REGION="us-central1"
SERVICE_NAME="brew-detective-backend"

echo "⚡ Quick Deploy - Forcing new Cloud Run revision..."

# Get current timestamp for forcing new revision
TIMESTAMP=$(date +%s)
echo "📅 Deployment timestamp: $TIMESTAMP"

# Method 1: Update deployment with timestamp annotation
echo "🔄 Updating service with new timestamp annotation..."
gcloud run services update $SERVICE_NAME \
    --region=$REGION \
    --project=$PROJECT_ID \
    --update-annotations="run.googleapis.com/last-deploy=$TIMESTAMP"

echo "✅ Quick deployment complete!"

# Check service status
echo "📊 Current service status:"
gcloud run services describe $SERVICE_NAME \
    --region=$REGION \
    --project=$PROJECT_ID \
    --format="table(metadata.name,status.url,status.conditions[0].type,status.conditions[0].status)"

echo ""
echo "🌐 Service URL: https://brew-detective-backend-1087966598090.us-central1.run.app"