FROM golang:1.26-bookworm AS build

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/blacklist-service ./cmd/blacklist-service

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/blacklist-service /app/blacklist-service

ENV SERVICE_NAME=blacklist-service \
    SERVER_PORT=8084

EXPOSE 8084

ENTRYPOINT ["/app/blacklist-service"]
