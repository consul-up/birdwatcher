# Birdwatcher - GCP Deployment Guide

This guide explains how to deploy the Birdwatcher application to Google Cloud Platform (GCP) using Cloud Run.

## Architecture

The application consists of two services:

- **Backend**: Go service that serves bird data via REST API (port 7000)
- **Frontend**: Go web application that displays bird information (port 6060)

Both services are deployed as containerized applications on Google Cloud Run, which provides:
- Automatic scaling (scales to zero when not in use)
- Pay-per-use pricing
- Built-in HTTPS
- Easy management

## Prerequisites

1. **Google Cloud Account**: Create one at [cloud.google.com](https://cloud.google.com)

2. **GCP Project**: Create a new project or use an existing one
   ```bash
   gcloud projects create birdwatcher-prod --name="Birdwatcher"
   gcloud config set project birdwatcher-prod
   ```

3. **Install Google Cloud SDK**: Download from [cloud.google.com/sdk](https://cloud.google.com/sdk/docs/install)

4. **Authenticate with GCP**:
   ```bash
   gcloud auth login
   gcloud auth configure-docker
   ```

5. **Enable Billing**: Make sure billing is enabled for your project in the GCP Console

## Deployment Options

You can deploy using either the **deployment script** (easiest) or **Terraform** (infrastructure as code).

### Option 1: Deployment Script (Recommended for Quick Setup)

The `deploy-to-gcp.sh` script handles the entire deployment process automatically.

#### Step 1: Set your GCP Project ID

```bash
export GCP_PROJECT_ID="your-gcp-project-id"
```

Optionally, you can also set:
```bash
export GCP_REGION="us-central1"  # Default region
export VERSION="v1"               # Backend version (v1 or v2)
```

#### Step 2: Run the deployment script

```bash
./deploy-to-gcp.sh
```

The script will:
1. Enable required GCP APIs (Cloud Run, Cloud Build, Container Registry)
2. Build and push Docker images to Google Container Registry
3. Deploy the backend service to Cloud Run
4. Deploy the frontend service with the backend URL
5. Output the URLs for both services

#### Step 3: Access Your Application

After deployment completes, you'll see output like:

```
=== Deployment Complete! ===

Frontend URL: https://birdwatcher-frontend-xxxxx.run.app
Backend URL: https://birdwatcher-backend-xxxxx.run.app

You can now access your application at:
  https://birdwatcher-frontend-xxxxx.run.app
```

Open the Frontend URL in your browser to use the application!

### Option 2: Terraform Deployment (Infrastructure as Code)

For production deployments, Terraform provides better control and repeatability.

#### Step 1: Install Terraform

Download from [terraform.io/downloads](https://www.terraform.io/downloads)

#### Step 2: Build and Push Docker Images

First, build and push the images manually:

```bash
# Set your project ID
export PROJECT_ID="your-gcp-project-id"

# Build and push backend
gcloud builds submit ./backend --tag="gcr.io/$PROJECT_ID/birdwatcher-backend"

# Build and push frontend
gcloud builds submit ./frontend --tag="gcr.io/$PROJECT_ID/birdwatcher-frontend"
```

#### Step 3: Configure Terraform

```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars` and set your values:
```hcl
project_id = "your-gcp-project-id"
region     = "us-central1"
backend_version = "v1"
```

#### Step 4: Deploy with Terraform

```bash
terraform init
terraform plan
terraform apply
```

Type `yes` when prompted to create the resources.

#### Step 5: Get the URLs

After deployment:
```bash
terraform output frontend_url
terraform output backend_url
```

## Configuration

### Environment Variables

#### Backend Service
- `VERSION`: Set to `v1` (birds) or `v2` (canaries) - default: `v1`
- `BIND_ADDR`: Listening address - default: `0.0.0.0:7000`
- `TRACING_URL`: Optional Zipkin tracing URL
- `GIN_MODE`: Gin framework mode - set to `release` in production

#### Frontend Service
- `BACKEND_URL`: Backend service URL (automatically configured)
- `BIND_ADDR`: Listening address - default: `0.0.0.0:6060`
- `TRACING_URL`: Optional Zipkin tracing URL
- `GIN_MODE`: Gin framework mode - set to `release` in production

### Updating Configuration

To update environment variables after deployment:

```bash
# Update backend version to v2 (canaries)
gcloud run services update birdwatcher-backend \
  --set-env-vars="VERSION=v2" \
  --region=us-central1

# Update frontend backend URL if needed
gcloud run services update birdwatcher-frontend \
  --set-env-vars="BACKEND_URL=https://new-backend-url.run.app" \
  --region=us-central1
```

## Monitoring and Logs

### View Logs

```bash
# Frontend logs
gcloud run logs read birdwatcher-frontend --region=us-central1

# Backend logs
gcloud run logs read birdwatcher-backend --region=us-central1

# Follow logs in real-time
gcloud run logs tail birdwatcher-frontend --region=us-central1
```

### View Metrics

Visit the [Cloud Run Console](https://console.cloud.google.com/run) to see:
- Request count
- Request latency
- Container instances
- Memory and CPU usage

## Scaling Configuration

Cloud Run auto-scales by default. To customize scaling:

```bash
# Set minimum and maximum instances
gcloud run services update birdwatcher-frontend \
  --min-instances=1 \
  --max-instances=10 \
  --region=us-central1
```

Note: Setting `min-instances > 0` keeps instances warm but incurs costs even with no traffic.

## Custom Domain

To use a custom domain:

1. **Verify domain ownership** in GCP Console
2. **Map domain to service**:
   ```bash
   gcloud run domain-mappings create \
     --service=birdwatcher-frontend \
     --domain=birdwatcher.yourdomain.com \
     --region=us-central1
   ```
3. **Update DNS** with the provided DNS records

## Cost Optimization

Cloud Run pricing is based on:
- Request count
- Compute time (CPU and memory)
- Networking

Tips to minimize costs:
1. Use default scaling (scales to zero when idle)
2. Set appropriate CPU and memory limits
3. Use Cloud CDN for static content
4. Monitor usage in the [Billing Console](https://console.cloud.google.com/billing)

## Updating the Application

### Using the Deployment Script

Just run the script again:
```bash
./deploy-to-gcp.sh
```

### Using Terraform

```bash
# Rebuild images
gcloud builds submit ./backend --tag="gcr.io/$PROJECT_ID/birdwatcher-backend"
gcloud builds submit ./frontend --tag="gcr.io/$PROJECT_ID/birdwatcher-frontend"

# Apply changes
cd terraform
terraform apply
```

### Manual Update

```bash
# Backend
gcloud builds submit ./backend --tag="gcr.io/$PROJECT_ID/birdwatcher-backend"
gcloud run deploy birdwatcher-backend \
  --image="gcr.io/$PROJECT_ID/birdwatcher-backend" \
  --region=us-central1

# Frontend
gcloud builds submit ./frontend --tag="gcr.io/$PROJECT_ID/birdwatcher-frontend"
gcloud run deploy birdwatcher-frontend \
  --image="gcr.io/$PROJECT_ID/birdwatcher-frontend" \
  --region=us-central1
```

## Cleanup

### Using the Deployment Script

The script output includes cleanup commands:
```bash
gcloud run services delete birdwatcher-frontend --region=us-central1
gcloud run services delete birdwatcher-backend --region=us-central1
```

### Using Terraform

```bash
cd terraform
terraform destroy
```

### Delete Container Images (Optional)

```bash
gcloud container images delete gcr.io/$PROJECT_ID/birdwatcher-backend
gcloud container images delete gcr.io/$PROJECT_ID/birdwatcher-frontend
```

## Troubleshooting

### Deployment Fails

1. **Check API enablement**:
   ```bash
   gcloud services list --enabled
   ```

2. **Verify authentication**:
   ```bash
   gcloud auth list
   ```

3. **Check project configuration**:
   ```bash
   gcloud config list
   ```

### Service Errors

1. **Check logs**:
   ```bash
   gcloud run logs read birdwatcher-frontend --region=us-central1 --limit=50
   ```

2. **Check service status**:
   ```bash
   gcloud run services describe birdwatcher-frontend --region=us-central1
   ```

3. **Test backend connectivity**:
   ```bash
   curl https://birdwatcher-backend-xxxxx.run.app/healthz
   ```

### Frontend Can't Reach Backend

The frontend needs the backend URL set correctly. Verify:
```bash
gcloud run services describe birdwatcher-frontend \
  --region=us-central1 \
  --format='value(spec.template.spec.containers[0].env[0].value)'
```

## CI/CD with Cloud Build

The included `cloudbuild.yaml` files enable automated deployments.

### Setup Triggers

1. Connect your repository to Cloud Build in the GCP Console
2. Create build triggers for backend and frontend
3. Push to your repository to trigger automatic deployments

### Manual Cloud Build

```bash
# Backend
gcloud builds submit ./backend --config=./backend/cloudbuild.yaml

# Frontend
gcloud builds submit ./frontend --config=./frontend/cloudbuild.yaml
```

## Security Best Practices

1. **Use Secret Manager** for sensitive data:
   ```bash
   echo -n "secret-value" | gcloud secrets create my-secret --data-file=-
   ```

2. **Restrict service access** (remove `--allow-unauthenticated` flag)

3. **Enable VPC connector** for private backend communication

4. **Use Cloud Armor** for DDoS protection

5. **Implement Cloud IAP** for authentication

## Support

For issues or questions:
- Check [Cloud Run documentation](https://cloud.google.com/run/docs)
- Review [pricing](https://cloud.google.com/run/pricing)
- See [Cloud Run quotas](https://cloud.google.com/run/quotas)

## Next Steps

After deployment:
1. Set up monitoring and alerting
2. Configure custom domain
3. Implement CI/CD pipelines
4. Set up staging environment
5. Configure backup and disaster recovery
