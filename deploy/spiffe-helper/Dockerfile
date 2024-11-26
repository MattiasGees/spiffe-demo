FROM --platform=${BUILDPLATFORM:-linux/amd64} ubuntu:24.04 AS tools-builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Download & unpack SPIFFE-helper
RUN apt-get update && apt-get install -y \
    wget 
RUN wget https://github.com/spiffe/spiffe-helper/releases/download/v0.7.0/spiffe-helper-v0.7.0.tar.gz
RUN tar zvxf spiffe-helper-v0.7.0.tar.gz
RUN chmod +x spiffe-helper

FROM --platform=${TARGETPLATFORM:-linux/amd64} ubuntu:24.04

RUN apt-get update && apt-get install -y \
    postgresql-client \
 && rm -rf /var/lib/apt/lists/*

# explicitly set user/group IDs
RUN set -eux; \
groupadd -r postgres --gid=999; \
# https://salsa.debian.org/postgresql/postgresql-common/blob/997d842ee744687d99a2b2d95c1083a2615c79e8/debian/postgresql-common.postinst#L32-35
useradd -r -g postgres --uid=999 --home-dir=/var/lib/postgresql --shell=/bin/bash postgres; \
# also create the postgres user's home directory with appropriate permissions
# see https://github.com/docker-library/postgres/issues/274
mkdir -p /var/lib/postgresql; \
chown -R postgres:postgres /var/lib/postgresql

USER 999

COPY --from=tools-builder /spiffe-helper /usr/bin/spiffe-helper

ENTRYPOINT ["/usr/bin/spiffe-helper"]
