# timo — build requires CGO (mattn/go-sqlite3)
FROM golang:1.22-bookworm AS build

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libc6-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/timo ./cmd/timo

# Test image (mount source at /src): docker compose --profile test run --rm test
# Must come before the final runtime stage so `docker build` defaults to the server image.
FROM golang:1.22-bookworm AS test
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc libc6-dev \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /src
ENV CGO_ENABLED=1
ENTRYPOINT ["go", "test", "-count=1", "./..."]

# --- runtime (default image target)
FROM debian:bookworm-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libsqlite3-0 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /out/timo /usr/local/bin/timo

WORKDIR /data

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/timo"]
CMD ["serve", "-addr", "0.0.0.0:8080", "-dsn", "file:/data/timo.db?cache=shared&_fk=1"]
