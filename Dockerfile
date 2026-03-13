FROM node:22-bookworm-slim AS frontend-builder

WORKDIR /src/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build


FROM golang:1.25-bookworm AS go-builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend-builder /src/pb_public ./pb_public

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/oana-sessions .


FROM debian:bookworm-slim AS runtime

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=go-builder /out/oana-sessions ./oana-sessions
COPY --from=go-builder /src/pb_public ./pb_public

VOLUME ["/app/pb_data"]

EXPOSE 8090

ENTRYPOINT ["./oana-sessions", "serve", "--http=0.0.0.0:8090"]
