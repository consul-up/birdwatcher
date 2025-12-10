# Birdwatcher - Quick Start Guide

Deploy to GCP in 3 simple steps!

## Prerequisites

- Google Cloud account
- gcloud CLI installed ([download here](https://cloud.google.com/sdk/docs/install))
- Billing enabled on your GCP project

## Quick Deploy

### 1. Set Your Project

```bash
export GCP_PROJECT_ID="your-project-id"
```

Or use your current gcloud project:
```bash
gcloud config set project your-project-id
```

### 2. Deploy

```bash
./deploy-to-gcp.sh
```

### 3. Done! 🎉

The script will output your application URL. Open it in your browser!

```
Frontend URL: https://birdwatcher-frontend-xxxxx.run.app
```

## What Gets Deployed?

- **Backend Service**: API serving bird data (Cloud Run)
- **Frontend Service**: Web UI for browsing birds (Cloud Run)
- Both services auto-scale based on traffic
- HTTPS enabled by default
- Estimated cost: ~$0-5/month for light usage (scales to zero when idle)

## Next Steps

- See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed configuration
- Monitor your app: [Cloud Console](https://console.cloud.google.com/run)
- View logs: `gcloud run logs tail birdwatcher-frontend`

## Clean Up

To remove all resources:

```bash
gcloud run services delete birdwatcher-frontend --region=us-central1
gcloud run services delete birdwatcher-backend --region=us-central1
```

## Troubleshooting

**APIs not enabled?**
```bash
gcloud services enable run.googleapis.com cloudbuild.googleapis.com
```

**Authentication issues?**
```bash
gcloud auth login
```

**Need help?** Check [DEPLOYMENT.md](DEPLOYMENT.md) for detailed troubleshooting.
