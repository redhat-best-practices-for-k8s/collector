ARG COLLECTOR_USER_UID=1000

#### Build executable binary ####
FROM golang:alpine AS builder

#hadolint ignore=DL3018
RUN apk update && apk add --no-cache git

ENV SRC_DIR=/tnf

ARG COLLECTOR_USER_UID
ARG COLLECTOR_USER=collectoruser

RUN adduser -D -u ${COLLECTOR_USER_UID} ${COLLECTOR_USER}

WORKDIR $SRC_DIR

COPY . .

# Fetch dependencies and Build the Go application
RUN go build

#### Build small image ####
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.9-1161

# Copy the user from the build image
COPY --from=builder /etc/passwd /etc/passwd

ARG COLLECTOR_USER_UID

USER ${COLLECTOR_USER_UID}

WORKDIR $SRC_DIR/collectoruser

# Copy the built app from the build image
COPY --from=builder /tnf/collector ./collector

EXPOSE 80

# Set the command to run when the container starts
ENTRYPOINT ["./collector"]
