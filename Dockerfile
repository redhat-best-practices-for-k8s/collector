ARG COLLECTOR_USER_UID=1000

#### Build executable binary ####
FROM golang:alpine AS builder

#hadolint ignore=DL3041
RUN apk update && apk add --no-cache git=2.40.1-r0

ENV SRC_DIR=/tnf

ARG COLLECTOR_USER_UID
ARG COLLECTOR_USER=collectoruser

RUN adduser -D -u ${COLLECTOR_USER_UID} ${COLLECTOR_USER}

WORKDIR $SRC_DIR

COPY . .

# Fetch dependencies and Build the Go application
RUN go build

#### Build small image ####
FROM ubi8-minimal:8.8-860

# Copy the user from the build image
COPY --from=builder /etc/passwd /etc/passwd

ARG COLLECTOR_USER_UID

USER ${COLLECTOR_USER_UID}

WORKDIR $SRC_DIR/collectoruser

# Copy the built app from the build image
COPY --from=builder /tnf/collector ./collector

EXPOSE 8080

# Set the command to run when the container starts
ENTRYPOINT ["./collector"]
