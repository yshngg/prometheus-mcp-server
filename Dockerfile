# syntax=docker/dockerfile:1

ARG VERSION_NUMBER=(unknown)

# Build stage
FROM --platform=$BUILDPLATFORM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION_NUMBER
ARG GIT_COMMIT
ARG BUILD_DATE
ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -o prometheus-mcp-server \
    -ldflags="-s -w -X 'github.com/yshngg/prometheus-mcp-server/internal/version.Number=${VERSION_NUMBER}' -X 'github.com/yshngg/prometheus-mcp-server/internal/version.GitCommit=${GIT_COMMIT}' -X 'github.com/yshngg/prometheus-mcp-server/internal/version.BuildDate=${BUILD_DATE}'" \
    .

# Final image
FROM alpine:3.23

LABEL org.opencontainers.image.source=https://github.com/yshngg/prometheus-mcp-server
LABEL org.opencontainers.image.description="A Prometheus Model Context Protocol Server."
LABEL org.opencontainers.image.licenses=Apache-2.0

WORKDIR /

COPY --from=builder /app/prometheus-mcp-server /prometheus-mcp-server

USER nobody

ENTRYPOINT ["/prometheus-mcp-server"]
