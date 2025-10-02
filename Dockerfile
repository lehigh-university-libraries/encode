FROM golang:1.25-alpine3.22@sha256:b6ed3fd0452c0e9bcdef5597f29cc1418f61672e9d3a2f55bf02e7222c014abd

SHELL ["/bin/ash", "-o", "pipefail", "-c"]

ARG \
  # renovate: datasource=repology depName=alpine_3_22/bash
  BASH_VERSION=5.2.37-r0 \
  # renovate: datasource=repology depName=alpine_3_22/ca-certificates
  CA_CERTIFICATES_VERSION=20250619-r0 \
  # renovate: datasource=repology depName=alpine_3_22/curl
  CURL_VERSION=8.14.1-r2 \
  # renovate: datasource=repology depName=alpine_3_22/openssl
  OPENSSL_VERSION=3.5.4-r0

WORKDIR /app

RUN adduser -S -G nobody encode

RUN apk update && \
    apk add --no-cache \
      bash=="${BASH_VERSION}" \
      ca-certificates=="${CA_CERTIFICATES_VERSION}" \
      curl=="${CURL_VERSION}" \
      openssl=="${OPENSSL_VERSION}"

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN chown -R encode:nobody /app /tmp

RUN go build -o /app/encode && \
  go clean -cache -modcache

ENTRYPOINT ["/app/encode"]
