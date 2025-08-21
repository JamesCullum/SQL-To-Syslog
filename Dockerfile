# Build Go
FROM golang:1 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /sql2syslog

# Run
FROM alpine:3 AS release-stage

RUN mkdir -p /app && chmod -R 777 /app && \
    apk update && apk upgrade --available && apk --no-cache add ca-certificates && update-ca-certificates
WORKDIR /app

COPY --from=build-stage /sql2syslog /app/sql2syslog
COPY config.json /app/config.json

ENTRYPOINT ["/app/sql2syslog"]
