#!/bin/bash

set -e

mkdir -p /tmp/aws

if [ "$AWS_AUTH" == "X509" ]; then
cat > /tmp/aws/config <<EOL
[default]
credential_process = /usr/bin/aws-spiffe-workload-helper x509-credential-process --profile-arn ${AWS_PROFILE_ARN} --trust-anchor-arn ${AWS_TRUST_ANCHOR_ARN} --role-arn ${AWS_ROLE_ARN}
EOL
else
cat > /tmp/aws/config <<EOL
[default]
credential_process = /usr/bin/spiffe-aws-assume-role credentials --role-arn ${AWS_ROLE_ARN} --audience ${JWT_AUDIENCE} --workload-socket ${SPIFFE_ENDPOINT_SOCKET}
EOL
fi