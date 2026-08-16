# syntax=docker/dockerfile:1

# ---- build ----
FROM golang:1.26-alpine AS build
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o /out/ddg .

# ---- certs ----
FROM alpine:latest AS certs
RUN apk add --no-cache ca-certificates

# ---- runtime ----
FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/ddg /ddg
USER 65534:65534
EXPOSE 8080
ENTRYPOINT ["/ddg", "--mcp-http", "8080"]
