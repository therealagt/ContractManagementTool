# Build from repository root:
#   docker build -t contract-api .
#   docker build --target extraction-worker -t contract-extraction-worker .
#   docker build --target archive-worker -t contract-archive-worker .
#   docker build --target integrity-cron -t contract-integrity-cron .

FROM golang:1.23-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY libs ./libs
COPY schemas ./schemas
COPY services/api ./services/api
COPY services/extraction-worker ./services/extraction-worker
COPY services/archive-worker ./services/archive-worker
COPY services/integrity-cron ./services/integrity-cron

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /api ./services/api/cmd/server
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /extraction-worker ./services/extraction-worker/cmd/worker
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /archive-worker ./services/archive-worker/cmd/worker
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /integrity-cron ./services/integrity-cron/cmd/worker

FROM gcr.io/distroless/static-debian12:nonroot AS api

COPY --from=builder /api /api
COPY --from=builder /src/schemas /schemas

EXPOSE 8080
ENTRYPOINT ["/api"]

FROM gcr.io/distroless/static-debian12:nonroot AS extraction-worker

COPY --from=builder /extraction-worker /extraction-worker
COPY --from=builder /src/schemas /schemas

EXPOSE 8080
ENTRYPOINT ["/extraction-worker"]

FROM gcr.io/distroless/static-debian12:nonroot AS archive-worker

COPY --from=builder /archive-worker /archive-worker

EXPOSE 8080
ENTRYPOINT ["/archive-worker"]

FROM gcr.io/distroless/static-debian12:nonroot AS integrity-cron

COPY --from=builder /integrity-cron /integrity-cron

EXPOSE 8080
ENTRYPOINT ["/integrity-cron"]
