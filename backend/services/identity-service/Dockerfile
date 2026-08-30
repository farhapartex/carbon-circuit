FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY gen ./gen
COPY internal ./internal
COPY services/identity-service ./services/identity-service

ARG REVISION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X main.revision=${REVISION}" \
    -o /out/identity-service \
    ./services/identity-service/cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/identity-service /identity-service

USER nonroot:nonroot
EXPOSE 9091

ENTRYPOINT ["/identity-service"]
