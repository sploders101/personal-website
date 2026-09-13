FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder

WORKDIR /app

# Copy and download dependencies (cached layer)
COPY go.mod ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build a statically linked, production-optimized binary
ARG TARGETARCH
RUN CGO_ENABLED=0 GOARCH=$TARGETARCH go build \
    -ldflags="-s -w" \
    -o /app/main .

FROM alpine:3.20

# Add a non-root user for security
RUN adduser -D appuser

RUN mkdir /app
RUN chown appuser:appuser /app

USER appuser

WORKDIR /app

COPY --from=builder /app/main /app/main
COPY ./assets /app/assets
COPY ./templates /app/templates

EXPOSE 8080

CMD ["/app/main"]
