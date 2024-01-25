# -----------------------------------------------------------------------------
#  Build Stage
# -----------------------------------------------------------------------------
FROM golang:1.21.5-alpine as build

# Important:
#   Because this is a CGO enabled package, you are required to set it as 1.
ENV CGO_ENABLED=1
ENV GOOS=linux

ENV GOBIN=$HOME/build

RUN apk add --no-cache \
    # Important: required for go-sqlite3
    gcc \
    # Required for Alpine
    musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /build ./cmd/main.go

# -----------------------------------------------------------------------------
#  Main Stage
# -----------------------------------------------------------------------------
FROM alpine

EXPOSE 50051

COPY --from=build build /build
COPY --from=build /app/config/docker_config.yaml /config/docker_config.yaml

CMD [ "/build", "--config=/config/docker_config.yaml" ]