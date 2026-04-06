# ── Build React frontend ──
FROM node:22-alpine AS frontend
WORKDIR /app
COPY dashboard/package*.json ./
RUN npm ci
COPY dashboard/ .
RUN npm run build

# ── Build Go binary ──
FROM golang:1.25 AS builder
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X 'github.com/kymaroshq/kymaros/internal/api.Version=${VERSION}'" -o kymaros ./cmd/main.go

# ── Runtime ──
FROM gcr.io/distroless/static:nonroot
LABEL org.opencontainers.image.source="https://github.com/kymaroshq/kymaros" \
      org.opencontainers.image.title="kymaros" \
      org.opencontainers.image.description="Continuous backup restore validation for Kubernetes" \
      org.opencontainers.image.licenses="Apache-2.0"
WORKDIR /
COPY --from=builder /workspace/kymaros .
COPY --from=frontend /app/dist ./static/
USER 65532:65532
EXPOSE 8080 8081 8443
ENTRYPOINT ["/kymaros"]
