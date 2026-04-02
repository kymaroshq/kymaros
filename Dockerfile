# Operator-only image (no dashboard)
FROM golang:1.25 AS builder
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG SIGNING_KEY=""
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a \
    -ldflags "-X github.com/kymaroshq/kymaros/internal/license.verifyKey=${SIGNING_KEY}" \
    -o manager cmd/main.go

FROM gcr.io/distroless/static:nonroot
LABEL org.opencontainers.image.source="https://github.com/kymaroshq/kymaros" \
      org.opencontainers.image.title="kymaros" \
      org.opencontainers.image.description="Kymaros Operator — continuous backup restore validation" \
      org.opencontainers.image.licenses="Apache-2.0"
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532
ENTRYPOINT ["/manager"]
