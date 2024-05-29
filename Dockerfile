FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.22.3 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG APPPATH

ARG GOPROXY

WORKDIR /workspace

COPY . .

RUN GOPROXY=$GOPROXY go mod download

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o bin/app ${APPPATH}/main.go

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine AS certs-builder

# Install ca-certificates package
RUN apk --update add ca-certificates

FROM --platform=${TARGETPLATFORM:-linux/amd64} scratch

WORKDIR /
USER 1001
COPY --from=builder /workspace/bin/app /usr/bin/app
COPY --from=certs-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/usr/bin/app"]
