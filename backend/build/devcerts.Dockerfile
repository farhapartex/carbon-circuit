FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /devcerts ./cmd/devcerts

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /devcerts /devcerts
ENTRYPOINT ["/devcerts"]
