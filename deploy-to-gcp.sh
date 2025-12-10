#!/bin/bash
set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID="${GCP_PROJECT_ID:-}"
REGION="${GCP_REGION:-us-central1}"
VERSION="${VERSION:-v1}"

echo -e "${GREEN}=== Birdwatcher GCP Deployment ===${NC}"
echo ""

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}Error: gcloud CLI is not installed${NC}"
    echo "Please install it from: https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Check if PROJECT_ID is set
if [ -z "$PROJECT_ID" ]; then
    echo -e "${YELLOW}GCP_PROJECT_ID not set. Attempting to get from gcloud config...${NC}"
    PROJECT_ID=$(gcloud config get-value project 2>/dev/null)

    if [ -z "$PROJECT_ID" ]; then
        echo -e "${RED}Error: No GCP project ID found${NC}"
        echo "Please set GCP_PROJECT_ID environment variable or configure gcloud:"
        echo "  export GCP_PROJECT_ID=your-project-id"
        echo "  or"
        echo "  gcloud config set project your-project-id"
        exit 1
    fi
fi

echo -e "${GREEN}Project ID:${NC} $PROJECT_ID"
echo -e "${GREEN}Region:${NC} $REGION"
echo -e "${GREEN}Version:${NC} $VERSION"
echo ""

# Enable required APIs
echo -e "${YELLOW}Enabling required GCP APIs...${NC}"
gcloud services enable run.googleapis.com \
    cloudbuild.googleapis.com \
    containerregistry.googleapis.com \
    --project="$PROJECT_ID"

echo ""
echo -e "${GREEN}Step 1: Building and deploying backend service...${NC}"

# Build and push backend image
gcloud builds submit ./backend \
    --tag="gcr.io/$PROJECT_ID/birdwatcher-backend" \
    --project="$PROJECT_ID"

# Deploy backend to Cloud Run
gcloud run deploy birdwatcher-backend \
    --image="gcr.io/$PROJECT_ID/birdwatcher-backend" \
    --platform=managed \
    --region="$REGION" \
    --allow-unauthenticated \
    --port=7000 \
    --set-env-vars="VERSION=$VERSION" \
    --project="$PROJECT_ID"

# Get backend URL
BACKEND_URL=$(gcloud run services describe birdwatcher-backend \
    --platform=managed \
    --region="$REGION" \
    --format='value(status.url)' \
    --project="$PROJECT_ID")

echo ""
echo -e "${GREEN}Backend deployed successfully!${NC}"
echo -e "${GREEN}Backend URL:${NC} $BACKEND_URL"
echo ""

echo -e "${GREEN}Step 2: Building and deploying frontend service...${NC}"

# Build and push frontend image
gcloud builds submit ./frontend \
    --tag="gcr.io/$PROJECT_ID/birdwatcher-frontend" \
    --project="$PROJECT_ID"

# Deploy frontend to Cloud Run
gcloud run deploy birdwatcher-frontend \
    --image="gcr.io/$PROJECT_ID/birdwatcher-frontend" \
    --platform=managed \
    --region="$REGION" \
    --allow-unauthenticated \
    --port=6060 \
    --set-env-vars="BACKEND_URL=$BACKEND_URL" \
    --project="$PROJECT_ID"

# Get frontend URL
FRONTEND_URL=$(gcloud run services describe birdwatcher-frontend \
    --platform=managed \
    --region="$REGION" \
    --format='value(status.url)' \
    --project="$PROJECT_ID")

echo ""
echo -e "${GREEN}=== Deployment Complete! ===${NC}"
echo ""
echo -e "${GREEN}Frontend URL:${NC} $FRONTEND_URL"
echo -e "${GREEN}Backend URL:${NC} $BACKEND_URL"
echo ""
echo -e "${YELLOW}You can now access your application at:${NC}"
echo -e "  ${FRONTEND_URL}"
echo ""
echo -e "${YELLOW}To view logs:${NC}"
echo -e "  gcloud run logs read birdwatcher-frontend --region=$REGION --project=$PROJECT_ID"
echo -e "  gcloud run logs read birdwatcher-backend --region=$REGION --project=$PROJECT_ID"
echo ""
echo -e "${YELLOW}To clean up resources:${NC}"
echo -e "  gcloud run services delete birdwatcher-frontend --region=$REGION --project=$PROJECT_ID"
echo -e "  gcloud run services delete birdwatcher-backend --region=$REGION --project=$PROJECT_ID"
echo ""
