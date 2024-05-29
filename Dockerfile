FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.22.3 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ARG GOPROXY

WORKDIR /workspace

COPY . .

RUN GOPROXY=$GOPROXY go mod download

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o bin/spiffe-demo

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine AS certs-builder

# Install ca-certificates package
RUN apk --update add ca-certificates

FROM --platform=${TARGETPLATFORM:-linux/amd64} scratch

WORKDIR /
USER 1001
COPY --from=builder /workspace/bin/spiffe-demo /usr/bin/spiffe-demo
COPY --from=certs-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/usr/bin/spiffe-demo"]
