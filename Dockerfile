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

FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine AS tools-builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Install ca-certificates package
RUN apk --update add ca-certificates

# Download & unpack AWS Assume Helper
RUN wget https://github.com/MattiasGees/spiffe-aws-assume-role/releases/download/v0.0.1-alpha2/spiffe-aws-assume-role-v0.0.1-alpha2-${TARGETOS}-${TARGETARCH}.tar.gz 
RUN tar zvxf spiffe-aws-assume-role-v0.0.1-alpha2-${TARGETOS}-${TARGETARCH}.tar.gz 

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine

WORKDIR /
USER 1001

COPY --from=builder /workspace/bin/spiffe-demo /usr/bin/spiffe-demo
COPY --from=tools-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=tools-builder /spiffe-aws-assume-role /usr/bin/spiffe-aws-assume-role

ENTRYPOINT ["/usr/bin/spiffe-demo"]
