FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.22.3 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ARG GOPROXY

WORKDIR /workspace

RUN apt-get update && apt-get install -y \
    wget 
RUN wget https://github.com/GoogleCloudPlatform/professional-services/archive/refs/heads/main.tar.gz
RUN tar zvxf main.tar.gz
 
RUN cd professional-services-main/tools/spiffe-gcp-proxy && GOPROXY=$GOPROXY go mod download

# Build
RUN cd professional-services-main/tools/spiffe-gcp-proxy && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o bin/spiffe-gcp-proxy

FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine AS tools-builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Install ca-certificates package
RUN apk --update add ca-certificates

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine

WORKDIR /
USER 1001

COPY --from=builder /workspace/professional-services-main/tools/spiffe-gcp-proxy/bin/spiffe-gcp-proxy /usr/bin/spiffe-gcp-proxy
COPY --from=tools-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/usr/bin/spiffe-gcp-proxy"]