#!/bin/bash

set -e

mkdir -p /tmp/aws

cat > /tmp/aws/config <<EOL
[default]
credential_process = /usr/bin/spiffe-aws-assume-role credentials --role-arn ${AWS_ROLE_ARN} --audience ${JWT_AUDIENCE} --workload-socket ${SPIFFE_ENDPOINT_SOCKET} --spiffe-id ${SPIFFE_ID}
EOL
