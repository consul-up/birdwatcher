# 🐦 Birdwatcher

A Go-based web application for exploring bird species with a microservices architecture.

## Overview

Birdwatcher is a full-stack application consisting of:

- **Backend**: RESTful API service that provides bird data and metadata
- **Frontend**: Web interface for browsing and viewing bird information

The application demonstrates:
- Microservices architecture
- Containerized deployment
- Cloud-native patterns
- Distributed tracing support
- Health monitoring

## Features

- 📖 Browse comprehensive bird species information
- 🎨 Beautiful web interface with bird images
- 🔄 Multiple bird datasets (standard birds and canaries)
- 📊 Built-in health checks and monitoring
- 🎯 Optional distributed tracing with Zipkin
- ⚡ Configurable synthetic delays for testing
- 🎲 Configurable error rates for resilience testing

## Quick Start

### Local Development

1. **Start the backend**:
   ```bash
   cd backend
   go run .
   # Backend runs on http://localhost:7000
   ```

2. **Start the frontend**:
   ```bash
   cd frontend
   go run .
   # Frontend runs on http://localhost:6060
   ```

3. **Open your browser**:
   ```
   http://localhost:6060
   ```

### Deploy to GCP

See [QUICKSTART.md](QUICKSTART.md) for a 3-step deployment to Google Cloud Platform.

```bash
export GCP_PROJECT_ID="your-project-id"
./deploy-to-gcp.sh
```

For detailed deployment options, see [DEPLOYMENT.md](DEPLOYMENT.md).

## Architecture

```
┌─────────────┐         ┌─────────────┐
│   Frontend  │ ──────▶ │   Backend   │
│   (Go/Gin)  │         │   (Go/Gin)  │
│   Port 6060 │         │   Port 7000 │
└─────────────┘         └─────────────┘
      │                        │
      │                        │
      ▼                        ▼
┌──────────────────────────────────┐
│        Static Assets &           │
│        Templates                 │
└──────────────────────────────────┘
```

## API Endpoints

### Backend (`http://localhost:7000`)

- `GET /bird` - Get a random bird
  - Query params:
    - `delay=<seconds>` - Add synthetic delay
    - `error-rate=<0-100>` - Simulate random errors
- `GET /healthz` - Health check endpoint

### Frontend (`http://localhost:6060`)

- `GET /` - Main application page
- `GET /admin` - Admin interface
- `GET /healthz` - Health check endpoint

## Configuration

### Backend Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `BIND_ADDR` | Server listen address | `0.0.0.0:7000` |
| `VERSION` | Bird dataset version (`v1` or `v2`) | `v1` |
| `TRACING_URL` | Zipkin tracing URL | (disabled) |
| `GIN_MODE` | Gin mode (`debug` or `release`) | `debug` |

### Frontend Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `BIND_ADDR` | Server listen address | `0.0.0.0:6060` |
| `BACKEND_URL` | Backend service URL | `http://localhost:7000` |
| `TRACING_URL` | Zipkin tracing URL | (disabled) |
| `GIN_MODE` | Gin mode (`debug` or `release`) | `debug` |

## Building

### Build Binaries

```bash
make build-bins
```

Builds binaries for multiple platforms:
- Linux (amd64, arm64)
- macOS (amd64, arm64)

Output: `bin/` directory

### Build Docker Images

```bash
# Backend
docker build -t birdwatcher-backend ./backend

# Frontend
docker build -t birdwatcher-frontend ./frontend
```

### Run with Docker

```bash
# Backend
docker run -p 7000:7000 birdwatcher-backend

# Frontend
docker run -p 6060:6060 \
  -e BACKEND_URL=http://host.docker.internal:7000 \
  birdwatcher-frontend
```

## Development

### Prerequisites

- Go 1.17 or higher
- Docker (optional, for containerized deployment)
- Make (optional, for using Makefile)

### Project Structure

```
birdwatcher/
├── backend/
│   ├── main.go           # Backend entry point
│   ├── birds.go          # Bird data handling
│   ├── tracing.go        # Distributed tracing
│   ├── birds.json        # Bird dataset v1
│   ├── canaries.json     # Bird dataset v2
│   ├── Dockerfile        # Backend container
│   └── cloudbuild.yaml   # GCP Cloud Build config
├── frontend/
│   ├── main.go           # Frontend entry point
│   ├── routes.go         # HTTP routes and handlers
│   ├── tracing.go        # Distributed tracing
│   ├── static/           # Static assets
│   ├── templates/        # HTML templates
│   ├── Dockerfile        # Frontend container
│   └── cloudbuild.yaml   # GCP Cloud Build config
├── terraform/            # Infrastructure as Code
│   ├── main.tf           # Main Terraform config
│   ├── variables.tf      # Variable definitions
│   └── outputs.tf        # Output definitions
├── deploy-to-gcp.sh      # GCP deployment script
├── Makefile              # Build automation
├── DEPLOYMENT.md         # Detailed deployment guide
├── QUICKSTART.md         # Quick deployment guide
└── README.md             # This file
```

### Testing

Test endpoints locally:

```bash
# Health check
curl http://localhost:7000/healthz

# Get a bird
curl http://localhost:7000/bird

# Test with delay
curl "http://localhost:7000/bird?delay=2"

# Test with error rate
curl "http://localhost:7000/bird?error-rate=50"
```

## Deployment Options

### Option 1: Quick Deploy to GCP
```bash
./deploy-to-gcp.sh
```

### Option 2: Terraform (Infrastructure as Code)
```bash
cd terraform
terraform init
terraform apply
```

### Option 3: Manual Docker Deployment
```bash
docker-compose up -d  # (if you create a docker-compose.yaml)
```

See [DEPLOYMENT.md](DEPLOYMENT.md) for comprehensive deployment instructions.

## Monitoring

### Health Checks

Both services expose `/healthz` endpoints:
```bash
curl http://localhost:6060/healthz
curl http://localhost:7000/healthz
```

### Distributed Tracing

Enable Zipkin tracing:

```bash
# Start Zipkin (example)
docker run -d -p 9411:9411 openzipkin/zipkin

# Run with tracing enabled
TRACING_URL=http://localhost:9411/api/v2/spans go run .
```

### Logs

View logs when running locally:
```bash
# Backend logs
cd backend && go run .

# Frontend logs
cd frontend && go run .
```

View logs on GCP:
```bash
gcloud run logs tail birdwatcher-frontend
gcloud run logs tail birdwatcher-backend
```

## License

This is a demonstration application.

## Contributing

This is a learning/demo project. Feel free to fork and experiment!

## Support

- GCP Deployment: See [DEPLOYMENT.md](DEPLOYMENT.md)
- Quick Start: See [QUICKSTART.md](QUICKSTART.md)
- Issues: Open an issue in your repository

## Acknowledgments

Bird data sourced from Wikipedia.
