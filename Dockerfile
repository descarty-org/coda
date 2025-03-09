# syntax=docker/dockerfile:1.4

# Stage 1: Build
FROM golang:1.24-alpine as builder

ARG VERSION
ARG TARGETOS
ARG TARGETARCH

ENV HOST=0.0.0.0
ENV PORT=8080

WORKDIR /app

RUN apk add --no-cache make

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /go/bin/coda ./cmd/coda

# Stage 2: Certs
FROM alpine:3.8 as certs
RUN apk --update add ca-certificates

# Stage 3: Final Image
FROM scratch AS final
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /app

COPY --from=builder /go/bin/coda /app/coda
COPY --from=builder /app/config /app/config

ENTRYPOINT [ "/app/coda" ]

EXPOSE 8080