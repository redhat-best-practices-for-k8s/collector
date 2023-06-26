#### Build executable binary ####
FROM golang:alpine AS builder

# hadolint ignore=DL3041
RUN apk update && apk add --no-cache git

ENV SRC_DIR=/tnf

ENV COLLECTOR_USER_UID=1000
ARG COLLECTOR_USER=collectoruser

RUN adduser -D -u "$COLLECTOR_USER_UID" $COLLECTOR_USER

WORKDIR $SRC_DIR

COPY . .

# Fetch dependencies and Build the Go application
RUN go get -d -v && go build

#### Build small image ####
FROM alpine:3.18

ENV COLLECTOR_USER_UID=$COLLECTOR_USER_UID

USER $COLLECTOR_USER_UID

WORKDIR $SRC_DIR/collectoruser

COPY --from=builder /tnf/collector ./collector

EXPOSE 8080

# Set the command to run when the container starts
ENTRYPOINT ["./collector"]
