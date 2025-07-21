#!/bin/bash

# Check Deployment Status Script
# Verifies the current deployment state and configuration

set -e

PROJECT_ID="brew-detective"
REGION="us-central1"
SERVICE_NAME="brew-detective-backend"

echo "🔍 Checking Brew Detective Backend Deployment Status..."
echo "📋 Project: $PROJECT_ID"
echo "🌍 Region: $REGION"
echo "🚀 Service: $SERVICE_NAME"
echo ""

# Check if user is authenticated
echo "👤 Checking authentication..."
if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" > /dev/null 2>&1; then
    echo "❌ Error: Not authenticated with gcloud"
    echo "Please run: gcloud auth login"
    exit 1
fi

ACTIVE_ACCOUNT=$(gcloud auth list --filter=status:ACTIVE --format="value(account)")
echo "✅ Authenticated as: $ACTIVE_ACCOUNT"
echo ""

# Check current gcloud project (for info only)
echo "📋 Current gcloud configuration:"
CURRENT_PROJECT=$(gcloud config get-value project 2>/dev/null || echo "Not set")
echo "   Current project: $CURRENT_PROJECT"
echo "   Target project: $PROJECT_ID"
if [ "$CURRENT_PROJECT" != "$PROJECT_ID" ]; then
    echo "ℹ️  Note: Using --project flag, so current config doesn't matter"
fi
echo ""

# Check if the service exists
echo "🔍 Checking if service exists..."
if gcloud run services describe $SERVICE_NAME --region=$REGION --project=$PROJECT_ID > /dev/null 2>&1; then
    echo "✅ Service exists"
    
    # Get service details
    echo ""
    echo "📊 Service Details:"
    gcloud run services describe $SERVICE_NAME \
        --region=$REGION \
        --project=$PROJECT_ID \
        --format="table(metadata.name,status.url,status.conditions[0].type,status.conditions[0].status)"
    
    echo ""
    echo "🔗 Service URL:"
    SERVICE_URL=$(gcloud run services describe $SERVICE_NAME --region=$REGION --project=$PROJECT_ID --format='value(status.url)')
    echo "   $SERVICE_URL"
    
    echo ""
    echo "📝 Recent Revisions:"
    gcloud run revisions list \
        --service=$SERVICE_NAME \
        --region=$REGION \
        --project=$PROJECT_ID \
        --limit=3 \
        --format="table(metadata.name,status.conditions[0].type,spec.containers[0].image,metadata.creationTimestamp)"
    
else
    echo "❌ Service does not exist"
    echo "To deploy: ./deploy.sh"
fi

echo ""
echo "🎯 Deployment Commands:"
echo "   Full deploy:  ./deploy.sh"
echo "   Force deploy: ./deploy.sh --force"
echo "   Quick deploy: ./quick-deploy.sh"
echo "   Check status: ./check-deployment.sh"