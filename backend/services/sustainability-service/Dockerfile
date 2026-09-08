FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY gen ./gen
COPY internal ./internal
COPY services/sustainability-service ./services/sustainability-service

ARG REVISION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w" \
    -o /out/sustainability-migrate \
    ./services/sustainability-service/cmd/migrate

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X main.revision=${REVISION}" \
    -o /out/sustainability-service \
    ./services/sustainability-service/cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/sustainability-migrate /sustainability-migrate
COPY --from=build /out/sustainability-service /sustainability-service

USER nonroot:nonroot
EXPOSE 9096

ENTRYPOINT ["/sustainability-service"]
