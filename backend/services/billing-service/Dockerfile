FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY gen ./gen
COPY internal ./internal
COPY services/billing-service ./services/billing-service

ARG REVISION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w" \
    -o /out/billing-migrate \
    ./services/billing-service/cmd/migrate

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X main.revision=${REVISION}" \
    -o /out/billing-service \
    ./services/billing-service/cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/billing-migrate /billing-migrate
COPY --from=build /out/billing-service /billing-service

USER nonroot:nonroot
EXPOSE 9092

ENTRYPOINT ["/billing-service"]
