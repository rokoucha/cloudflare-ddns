# syntax=docker/dockerfile:1.23

FROM docker.io/library/golang:1.26 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/cloudflare-ddns .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/cloudflare-ddns /cloudflare-ddns

ENTRYPOINT ["/cloudflare-ddns"]
