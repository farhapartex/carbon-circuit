FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY gen ./gen
COPY internal ./internal
COPY services/provenance-read-service ./services/provenance-read-service

ARG REVISION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w" \
    -o /out/provenance-read-migrate \
    ./services/provenance-read-service/cmd/migrate

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X main.revision=${REVISION}" \
    -o /out/provenance-read-service \
    ./services/provenance-read-service/cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/provenance-read-migrate /provenance-read-migrate
COPY --from=build /out/provenance-read-service /provenance-read-service

USER nonroot:nonroot
EXPOSE 9094

ENTRYPOINT ["/provenance-read-service"]
