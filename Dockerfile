FROM docker.io/library/golang:1.19-bullseye AS build

WORKDIR /app

COPY go.mod /app/go.mod
COPY go.sum /app/go.sum
RUN go mod download

COPY *.go /app/

RUN go build -o /app/cloudflare-ddns

FROM docker.io/library/debian:bullseye-slim

ARG USER_NAME="cloudflare-ddns"
ARG USER_ID="999"

RUN useradd -l -u "${USER_ID}" -m "${USER_NAME}"

RUN apt update \
    && apt full-upgrade -y \
    && apt install -y --no-install-recommends \
    ca-certificates \
    openssl

WORKDIR /app

COPY --from=build /app/cloudflare-ddns /app/cloudflare-ddns

USER $USER_NAME

ENTRYPOINT ["/app/cloudflare-ddns"]
