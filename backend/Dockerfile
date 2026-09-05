# Multi-stage; TARGETARCH vem do buildx (--platform linux/arm64 → arm64) para Fargate ARM.
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETARCH
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /api ./cmd/api

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /api .
# Asset seed VOD (migration 003_seed_vod_demo) — MP4 curto versionado no repo.
COPY assets/vod /app/assets/vod
# Fontes TTF para PDF de certificado (acentos PT); PDF gerado on-the-fly (sem volume).
COPY assets/certs /app/assets/certs

EXPOSE 8080
CMD ["./api"]
