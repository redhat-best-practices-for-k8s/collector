# Start with a base image
FROM golang:1.20

ENV GOPATH=/root/go
ENV SRC_DIR=/tnf
# Copy the source code into the container
COPY . $SRC_DIR

# Change to the workdir
WORKDIR $SRC_DIR

# Build the Go application
RUN go build

EXPOSE 8080

# Set the command to run when the container starts
CMD ["./collector"]
