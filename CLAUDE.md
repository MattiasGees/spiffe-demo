# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SPIFFE Demo is a Go application showcasing SPIFFE (Secure Production Identity Framework For Everyone) use-cases in production-like environments. It demonstrates secure service-to-service communication, cloud provider integration (AWS, GCP), and database authentication using SPIFFE identities.

## Build Commands

```bash
# Build all Docker images locally (requires docker buildx)
make build

# Publish multi-platform images to registry
make publish
```

No unit tests exist - this is a demonstration/example project.

## Architecture

The application has three subcommands that run as separate services:

1. **customer** - Entry point web service exposed via Ingress. Provides buttons to:
   - Connect to SPIFFE-authenticated backend (application-layer mTLS)
   - Connect to HTTP backend via Envoy (infrastructure-layer mTLS)
   - Read/write AWS S3 with SPIFFE JWT identity
   - Read/write Google Cloud Storage with SPIFFE JWT identity
   - Authenticate to PostgreSQL using X.509 certificates

2. **backend** - SPIFFE-native service using go-spiffe for mTLS server

3. **httpservice** - Plain HTTP service fronted by Envoy proxy for SPIFFE auth

### Key Integration Patterns

- **Application Layer**: Direct SPIFFE integration via `github.com/spiffe/go-spiffe/v2`
- **Infrastructure Layer**: Envoy sidecar handles SPIFFE mTLS
- **AWS Auth**: Uses `spiffe-aws-assume-role` via credential_process, or X.509 with AWS IAM Roles Anywhere
- **GCP Auth**: Uses `spiffe-gcp-proxy` sidecar to intercept metadata API calls
- **PostgreSQL**: Certificate-based auth via SPIFFE-helper sidecar

### Code Structure

- `cmd/` - CLI command definitions using cobra
- `pkg/customer/` - Customer service implementation (HTTP handlers for each integration)
- `pkg/backend/` - mTLS server with SPIFFE validation
- `pkg/httpservice/` - Plain HTTP server
- `deploy/chart/` - Helm chart for Kubernetes deployment
- `deploy/terraform/` - AWS and GCP infrastructure setup
- `deploy/` subdirectories - Dockerfiles for sidecars (postgresql, spiffe-helper, spiffe-gcp-proxy, initcontainer)

### SPIFFE Integration Points

```go
// Workload API connection
source, err := workloadapi.NewX509Source(ctx)

// mTLS client
tlsConfig := tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeID(serverID))

// mTLS server
tlsConfig := tlsconfig.MTLSServerConfig(source, source, tlsconfig.AuthorizeID(clientID))
```

## Deployment

Requires:
- Kubernetes cluster with Ingress controller
- cert-manager with `letsencrypt-prod` ClusterIssuer
- SPIRE server (install via Helm from spiffe.github.io/helm-charts-hardened/)

Deploy with Helm:
```bash
helm upgrade --install -n spiffe-demo spife-demo ./deploy/chart/spiffe-demo --create-namespace
```
