# Build Go
FROM golang:1.22 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /sql2syslog

# Run
FROM alpine:3 AS release-stage

RUN mkdir -p /app && chmod -R 777 /app
WORKDIR /app

COPY --from=build-stage /sql2syslog /app/sql2syslog

ENTRYPOINT ["/app/sql2syslog"]