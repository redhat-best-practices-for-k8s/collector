#### Build executable binary ####
FROM golang:alpine AS builder

ENV SRC_DIR=/tnf

RUN apk update && apk add --no-cache git

WORKDIR $SRC_DIR

COPY . .

# Fetch dependencies
RUN go get -d -v

# Build the Go application
RUN go build

#### Build small image ####
FROM alpine

ENV COLLECTOR_USER_UID=1000

RUN adduser -D -u "$COLLECTOR_USER_UID" collectoruser

USER collectoruser

WORKDIR $SRC_DIR/collectoruser

COPY --from=builder /tnf/collector ./collector

EXPOSE 8080

# Set the command to run when the container starts
ENTRYPOINT ["./collector"]
