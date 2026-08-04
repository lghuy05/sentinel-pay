FROM golang:1.26-bookworm AS build

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/account-service ./cmd/account-service

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/account-service /app/account-service

ENV SERVICE_NAME=account-service \
    SERVER_PORT=8087

EXPOSE 8087

ENTRYPOINT ["/app/account-service"]
