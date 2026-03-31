# =============================================================================
# Stage 1 — builder
# Compiles a fully static, stripped binary inside an Alpine-based Go toolchain.
# CGO_ENABLED=0 + -trimpath + -ldflags="-s -w" minimises binary size.
# =============================================================================
FROM golang:1.25-alpine AS builder

# Install git for VCS-stamped builds (go build reads git metadata)
RUN apk add --no-cache git

WORKDIR /build

# Cache dependency layer before copying source
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 go build \
      -trimpath \
      -ldflags="-s -w \
        -X github.com/0xc0d/vessel/cmd.Version=${VERSION} \
        -X github.com/0xc0d/vessel/cmd.Commit=${COMMIT} \
        -X github.com/0xc0d/vessel/cmd.BuildTime=${BUILD_TIME}" \
      -o vessel .

# Report binary size during build for observability
RUN ls -lh vessel && echo "Binary built successfully"

# =============================================================================
# Stage 2 — runtime
# Minimal Alpine image with only the kernel-level tools vessel needs:
#   iptables  — network namespace / bridge rules
#   iproute2  — ip link / ip addr commands for veth setup
# The Go binary is statically linked so no libc or Go runtime is needed.
# =============================================================================
FROM alpine:3.18

RUN apk add --no-cache iptables iproute2

# vessel stores container state and image layers in these paths
RUN mkdir -p /var/lib/vessel/images/layers \
             /var/run/vessel/containers \
             /var/run/vessel/netns

COPY --from=builder /build/vessel /usr/local/bin/vessel

# vessel requires root for namespace and cgroup operations
USER root

ENTRYPOINT ["/usr/local/bin/vessel"]
CMD ["--help"]
