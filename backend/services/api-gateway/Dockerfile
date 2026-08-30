FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY gen ./gen
COPY internal ./internal
COPY services/api-gateway ./services/api-gateway

ARG REVISION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X main.revision=${REVISION}" \
    -o /out/api-gateway \
    ./services/api-gateway/cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/api-gateway /api-gateway

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/api-gateway"]
