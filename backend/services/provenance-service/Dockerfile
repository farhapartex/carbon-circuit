FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY gen ./gen
COPY internal ./internal
COPY services/provenance-service ./services/provenance-service

ARG REVISION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w" \
    -o /out/provenance-migrate \
    ./services/provenance-service/cmd/migrate

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X main.revision=${REVISION}" \
    -o /out/provenance-service \
    ./services/provenance-service/cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/provenance-migrate /provenance-migrate
COPY --from=build /out/provenance-service /provenance-service

USER nonroot:nonroot
EXPOSE 9093

ENTRYPOINT ["/provenance-service"]
