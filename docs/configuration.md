# Configuration Reference

This document describes the configuration options for the SPIFFE Demo application.

## Config File Format (YAML)

```yaml
# config.yaml
# Global settings (apply to all services)
server:
  address: "0.0.0.0:8080"

spiffe:
  authorized_id: "spiffe://example.org/backend"

# Service-specific settings
customer:
  backend:
    service_url: "https://backend:9090"

  http_backend:
    service_url: "https://httpbackend:8080"
    spiffe_id: "spiffe://example.org/httpservice"

  aws:
    bucket: "my-s3-bucket"
    file_path: "testfile"
    # Note: AWS region is NOT configured here - use standard AWS_REGION
    # or AWS_DEFAULT_REGION environment variables (AWS SDK standard)

  gcp:
    bucket: "my-gcs-bucket"
    file_path: "testfile"
    proxy_url: "http://localhost:8081"  # spiffe-gcp-proxy sidecar URL

  postgresql:
    host: "postgres.local"
    user: "spiffe_user"

backend:
  # Uses only global settings (server.address, spiffe.authorized_id)

httpservice:
  # Uses only server.address
```

## Configuration Priority

**Priority order (highest to lowest):**
1. Command-line flags (explicit user intent)
2. Environment variables (container-friendly)
3. Config file
4. Defaults

## Environment Variable Mapping

| Config Key | Environment Variable | CLI Flag |
|------------|---------------------|----------|
| `server.address` | `SPIFFE_DEMO_SERVER_ADDRESS` | `--server-address` |
| `spiffe.authorized_id` | `SPIFFE_DEMO_SPIFFE_AUTHORIZED_ID` | `--authorized-spiffe` |
| `customer.backend.service_url` | `SPIFFE_DEMO_CUSTOMER_BACKEND_SERVICE_URL` | `--backend-service` |
| `customer.http_backend.service_url` | `SPIFFE_DEMO_CUSTOMER_HTTP_BACKEND_SERVICE_URL` | `--httpbackend-service` |
| `customer.http_backend.spiffe_id` | `SPIFFE_DEMO_CUSTOMER_HTTP_BACKEND_SPIFFE_ID` | `--authorized-spiffe-httpbackend` |
| `customer.aws.bucket` | `SPIFFE_DEMO_CUSTOMER_AWS_BUCKET` | `--aws-bucket` |
| `customer.aws.file_path` | `SPIFFE_DEMO_CUSTOMER_AWS_FILE_PATH` | `--aws-file-path` |
| `customer.gcp.bucket` | `SPIFFE_DEMO_CUSTOMER_GCP_BUCKET` | `--gcp-bucket` |
| `customer.gcp.file_path` | `SPIFFE_DEMO_CUSTOMER_GCP_FILE_PATH` | `--gcp-file-path` |
| `customer.gcp.proxy_url` | `SPIFFE_DEMO_CUSTOMER_GCP_PROXY_URL` | `--gcp-proxy-url` |
| `customer.postgresql.host` | `SPIFFE_DEMO_CUSTOMER_POSTGRESQL_HOST` | `--postgresql-host` |
| `customer.postgresql.user` | `SPIFFE_DEMO_CUSTOMER_POSTGRESQL_USER` | `--postgresql-user` |

## Example Usage

```bash
# Using config file
spiffe-demo customer --config /etc/spiffe-demo/config.yaml

# Using CLI flags
spiffe-demo customer \
  --authorized-spiffe "spiffe://example.org/backend" \
  --server-address "0.0.0.0:8080" \
  --backend-service "https://backend:9090"

# Mixed: config file + CLI override
spiffe-demo customer --config config.yaml --server-address "0.0.0.0:9090"

# Using environment variables
export SPIFFE_DEMO_SERVER_ADDRESS="0.0.0.0:8080"
export SPIFFE_DEMO_CUSTOMER_AWS_BUCKET="my-bucket"
export AWS_REGION="eu-west-2"  # Standard AWS SDK env var
spiffe-demo customer
```

## Helm Chart Configuration

The Helm chart uses a ConfigMap to provide configuration:

```yaml
volumes:
  - name: config
    configMap:
      name: spiffe-demo-config
volumeMounts:
  - name: config
    mountPath: /etc/spiffe-demo
args:
  - customer
env:
  - name: AWS_REGION
    value: "eu-west-2"
```
