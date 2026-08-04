FROM golang:1.26-bookworm AS build

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/feature-extractor ./cmd/feature-extractor

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/feature-extractor /app/feature-extractor

ENV SERVICE_NAME=feature-extractor \
    SERVER_PORT=8082

EXPOSE 8082

ENTRYPOINT ["/app/feature-extractor"]
