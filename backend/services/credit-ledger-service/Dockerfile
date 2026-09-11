FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY gen ./gen
COPY internal ./internal
COPY services/credit-ledger-service ./services/credit-ledger-service

ARG REVISION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w" \
    -o /out/credit-ledger-migrate \
    ./services/credit-ledger-service/cmd/migrate

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X main.revision=${REVISION}" \
    -o /out/credit-ledger-service \
    ./services/credit-ledger-service/cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/credit-ledger-migrate /credit-ledger-migrate
COPY --from=build /out/credit-ledger-service /credit-ledger-service

USER nonroot:nonroot
EXPOSE 9097

ENTRYPOINT ["/credit-ledger-service"]
