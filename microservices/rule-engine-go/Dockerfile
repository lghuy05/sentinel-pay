FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/rule-engine ./cmd/rule-engine

FROM gcr.io/distroless/static-debian12:nonroot

ENV SERVICE_NAME=rule-engine \
    SERVER_PORT=8083

COPY --from=build /out/rule-engine /rule-engine
EXPOSE 8083
USER nonroot:nonroot
ENTRYPOINT ["/rule-engine"]
