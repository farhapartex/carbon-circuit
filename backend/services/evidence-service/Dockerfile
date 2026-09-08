FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY gen ./gen
COPY internal ./internal
COPY services/evidence-service ./services/evidence-service

ARG REVISION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w" \
    -o /out/evidence-migrate \
    ./services/evidence-service/cmd/migrate

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X main.revision=${REVISION}" \
    -o /out/evidence-service \
    ./services/evidence-service/cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/evidence-migrate /evidence-migrate
COPY --from=build /out/evidence-service /evidence-service

USER nonroot:nonroot
EXPOSE 9095

ENTRYPOINT ["/evidence-service"]
