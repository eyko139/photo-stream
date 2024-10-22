# syntax=docker/dockerfile:1

FROM golang:1.21

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download
RUN mkdir ./web
RUN mkdir ./web/pics
RUN mkdir ./web/thumbs

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY ./ ./

# Build
RUN GOARCH=wasm GOOS=js go build -o web/app.wasm
RUN go build

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/reference/dockerfile/#expose
EXPOSE 8001

# Run
CMD ["./photo-stream"]
