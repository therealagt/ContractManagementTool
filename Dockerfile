# Build from repository root:
#   docker build -t contract-api .
FROM golang:1.23-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY libs ./libs
COPY services/api ./services/api

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /api ./services/api/cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /api /api

EXPOSE 8080
ENTRYPOINT ["/api"]
