FROM golang:1.21-alpine3.19

MAINTAINER Vicky Phang <vickyphang11@gmail.com>

# Set destination for COPY
WORKDIR /app


# Copy the source code
COPY src/archimedes/ ./

# Download Go modules
RUN go mod download

# Build
RUN go build -o /archimedes

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/reference/dockerfile/#expose
EXPOSE 5000

# Run
CMD ["/archimedes"]