#!/bin/sh
set -e

echo "Waiting for SPIRE Server to be ready..."
sleep 5

SOCKET="/opt/spire/sockets/api.sock"

echo "Creating registration entries..."

# Register the ledger service
# Uses docker workload attestor with compose service label
/opt/spire/bin/spire-server entry create \
    -socketPath "$SOCKET" \
    -spiffeID spiffe://example.org/ledger \
    -parentID spiffe://example.org/spire-agent \
    -selector docker:label:com.docker.compose.service:ledger \
    -x509SVIDTTL 3600 \
    -dns ledger \
    -dns localhost \
    2>/dev/null || echo "Ledger entry may already exist"

# Register the postgres spiffe-helper to get certs for PostgreSQL
# The CN will be extracted from the SPIFFE ID path component
/opt/spire/bin/spire-server entry create \
    -socketPath "$SOCKET" \
    -spiffeID spiffe://example.org/postgres \
    -parentID spiffe://example.org/spire-agent \
    -selector docker:label:com.docker.compose.service:postgres-spiffe-helper \
    -x509SVIDTTL 3600 \
    -dns postgres \
    -dns localhost \
    2>/dev/null || echo "Postgres entry may already exist"

# Register the test client (simulating customer service)
/opt/spire/bin/spire-server entry create \
    -socketPath "$SOCKET" \
    -spiffeID spiffe://example.org/customer \
    -parentID spiffe://example.org/spire-agent \
    -selector docker:label:com.docker.compose.service:test-client \
    -x509SVIDTTL 3600 \
    -dns test-client \
    -dns localhost \
    2>/dev/null || echo "Customer entry may already exist"

echo ""
echo "Listing all entries:"
/opt/spire/bin/spire-server entry show -socketPath "$SOCKET"

echo ""
echo "Registration complete!"
