FROM golang:1.26-bookworm AS build

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/fraud-orchestrator ./cmd/fraud-orchestrator

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/fraud-orchestrator /app/fraud-orchestrator

ENV SERVICE_NAME=fraud-orchestrator \
    SERVER_PORT=8085

EXPOSE 8085

ENTRYPOINT ["/app/fraud-orchestrator"]
