# ============================================
# Dockerfile — slimify
# Uses pre-built binary from GoReleaser
# ============================================

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY slimify /usr/local/bin/slimify

ENTRYPOINT ["slimify"]
