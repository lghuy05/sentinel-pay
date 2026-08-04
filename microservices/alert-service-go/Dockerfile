FROM golang:1.26-bookworm AS build

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/alert-service ./cmd/alert-service

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/alert-service /app/alert-service

ENV SERVICE_NAME=alert-service \
    SERVER_PORT=8086

EXPOSE 8086

ENTRYPOINT ["/app/alert-service"]
