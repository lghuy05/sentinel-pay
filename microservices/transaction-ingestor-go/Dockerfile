FROM golang:1.26-bookworm AS build

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/transaction-ingestor ./cmd/transaction-ingestor

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/transaction-ingestor /app/transaction-ingestor

ENV SERVICE_NAME=transaction-ingestor \
    SERVER_PORT=8081

EXPOSE 8081

ENTRYPOINT ["/app/transaction-ingestor"]
