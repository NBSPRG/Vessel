# syntax=docker/dockerfile:1.7
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
COPY main.go ./

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_TIME=unknown

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build \
      -trimpath \
      -ldflags="-s -w \
        -X github.com/0xc0d/vessel/cmd.Version=${VERSION} \
        -X github.com/0xc0d/vessel/cmd.Commit=${COMMIT} \
        -X github.com/0xc0d/vessel/cmd.BuildTime=${BUILD_TIME}" \
      -o /out/vessel .

FROM alpine:3.20

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_TIME=unknown

LABEL org.opencontainers.image.title="vessel" \
      org.opencontainers.image.description="Lightweight container runtime built in Go" \
      org.opencontainers.image.version=$VERSION \
      org.opencontainers.image.revision=$COMMIT \
      org.opencontainers.image.created=$BUILD_TIME

RUN apk add --no-cache ca-certificates iproute2 iptables \
    && mkdir -p /var/lib/vessel/images/layers \
                /var/run/vessel/containers \
                /var/run/vessel/netns

COPY --from=builder /out/vessel /usr/local/bin/vessel

USER root

ENTRYPOINT ["/usr/local/bin/vessel"]
CMD ["--help"]
