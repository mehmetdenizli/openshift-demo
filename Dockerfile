# ---- Build Stage ----
FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY app/go.mod ./
COPY app/main.go ./

RUN go build -ldflags="-s -w" -o demo-app .

# ---- Runtime Stage ----
FROM alpine:3.19

# OpenShift runners use random UIDs but are always members of the root group (GID 0)
RUN mkdir -p /data/app && \
    chgrp -R 0 /data/app && \
    chmod -R g=u /data/app

USER appuser
WORKDIR /app
COPY --from=builder /src/demo-app .

EXPOSE 8080
ENTRYPOINT ["./demo-app"]
