FROM docker.io/library/golang:1.19-alpine3.16 AS build

WORKDIR /app

COPY go.mod /app/go.mod
COPY go.sum /app/go.sum
RUN go mod download

COPY *.go /app/

RUN go build -o /app/cloudflare-ddns

FROM docker.io/library/alpine:3.16

ARG USER_NAME="cloudflare-ddns"
ARG USER_ID="998"

RUN adduser -u "${USER_ID}" -D "${USER_NAME}"

RUN apk add --no-cache \
    openssl ca-certificates

WORKDIR /app

COPY --from=build /app/cloudflare-ddns /app/cloudflare-ddns

USER $USER_NAME

ENTRYPOINT ["/app/cloudflare-ddns"]
