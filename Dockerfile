# ---- Build Stage ----
FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY app/go.mod ./
COPY app/main.go ./

RUN go build -ldflags="-s -w" -o demo-app .

# ---- Runtime Stage ----
FROM alpine:3.19

# OpenShift runners use random UIDs; file permissions must be set accordingly
RUN addgroup -S appgroup && adduser -S appuser -G appgroup \
    && mkdir -p /data/app \
    && chown -R appuser:appgroup /data/app \
    && chmod -R g+rwX /data/app   # Group write is required for OpenShift random-UID

USER appuser
WORKDIR /app
COPY --from=builder /src/demo-app .

EXPOSE 8080
ENTRYPOINT ["./demo-app"]
