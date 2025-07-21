#!/bin/bash

# Brew Detective Backend Deployment Script
set -e

# Parse command line arguments
FORCE_DEPLOY=false
SKIP_BUILD=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --force)
            FORCE_DEPLOY=true
            shift
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --force       Force a new deployment even if code hasn't changed"
            echo "  --skip-build  Skip Docker build and push, only update deployment"
            echo "  --help        Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# Configuration
PROJECT_ID="brew-detective"
REGION="us-central1"
SERVICE_NAME="brew-detective-backend"
IMAGE_BASE="us-central1-docker.pkg.dev/${PROJECT_ID}/brew-detective-repo/${SERVICE_NAME}"

# Generate unique tag based on timestamp and git commit (if available)
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
IMAGE_TAG="${TIMESTAMP}-${GIT_COMMIT}"
IMAGE_NAME="${IMAGE_BASE}:${IMAGE_TAG}"

echo "🚀 Starting deployment for Brew Detective Backend..."
echo "📁 Working from directory: ${SCRIPT_DIR}"
echo "🏷️  Image tag: ${IMAGE_TAG}"
echo "📦 Full image name: ${IMAGE_NAME}"

if [ "$FORCE_DEPLOY" = true ]; then
    echo "💪 Force deployment enabled - will create new revision regardless"
fi

if [ "$SKIP_BUILD" = true ]; then
    echo "⏭️  Build skip enabled - using existing Docker image"
fi

# Step 1: Build Docker image (unless skipped)
if [ "$SKIP_BUILD" = false ]; then
    echo "📦 Building Docker image for linux/amd64..."
    cd "${SCRIPT_DIR}"
    
    # Ensure Docker is authenticated for the correct project
    echo "🔐 Configuring Docker authentication..."
    gcloud auth configure-docker us-central1-docker.pkg.dev --project=${PROJECT_ID} --quiet
    
    docker build --platform linux/amd64 -t ${IMAGE_NAME} .

    # Step 2: Push to Artifact Registry
    echo "⬆️  Pushing image to Artifact Registry..."
    docker push ${IMAGE_NAME}
else
    echo "⏩ Skipping Docker build and push..."
fi

# Step 3: Prepare deployment manifest with timestamp
echo "🌐 Preparing deployment manifest..."
TIMESTAMP=$(date +%s)
echo "📅 Build timestamp: $TIMESTAMP"

# Create temporary deployment file with unique timestamp and image tag
sed -e "s/REPLACE_WITH_TIMESTAMP/$TIMESTAMP/g" \
    -e "s|FULL_IMAGE_NAME|${IMAGE_NAME}|g" \
    "${SCRIPT_DIR}/deploy.yaml" > "${SCRIPT_DIR}/deploy-tmp.yaml"

# Deploy to Cloud Run using manifest
echo "🚀 Deploying to Cloud Run..."
echo "📦 Using image: ${IMAGE_NAME}"
gcloud run services replace "${SCRIPT_DIR}/deploy-tmp.yaml" --region=${REGION} --project=${PROJECT_ID}

# Clean up temporary file
rm "${SCRIPT_DIR}/deploy-tmp.yaml"

# Step 4: Get the service URL
echo "✅ Deployment complete!"
SERVICE_URL=$(gcloud run services describe ${SERVICE_NAME} --region=${REGION} --project=${PROJECT_ID} --format='value(status.url)')
echo "🔗 Service URL: ${SERVICE_URL}"

echo "🎉 Backend deployment successful!"