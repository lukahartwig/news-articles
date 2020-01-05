FROM golang:1.13 AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM base AS spiegel-binary
WORKDIR /app/cmd/spiegel
RUN go build -o /usr/local/bin/spiegel

FROM base AS bild-binary
WORKDIR /app/cmd/bild
RUN go build -o /usr/local/bin/bild

FROM base AS russiatoday-binary
WORKDIR /app/cmd/rt
RUN go build -o /usr/local/bin/russiatoday

# hadolint ignore=DL3007
FROM gcr.io/distroless/base:latest AS spiegel
COPY --from=spiegel-binary /usr/local/bin/spiegel /usr/local/bin/spiegel
ENTRYPOINT ["/usr/local/bin/spiegel"]

# hadolint ignore=DL3007
FROM gcr.io/distroless/base:latest AS bild
COPY --from=bild-binary /usr/local/bin/bild /usr/local/bin/bild
ENTRYPOINT ["/usr/local/bin/bild"]

# hadolint ignore=DL3007
FROM gcr.io/distroless/base:latest AS russiatoday
COPY --from=russiatoday-binary /usr/local/bin/russiatoday /usr/local/bin/russiatoday
ENTRYPOINT ["/usr/local/bin/russiatoday"]