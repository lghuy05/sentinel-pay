FROM golang:1.26-bookworm AS build

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/api-gateway ./cmd/api-gateway

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/api-gateway /app/api-gateway

ENV SERVICE_NAME=api-gateway \
    SERVER_PORT=8080

EXPOSE 8080

ENTRYPOINT ["/app/api-gateway"]
