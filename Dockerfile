# Build stage
FROM golang:1.20.7-alpine3.17 AS build-env
# Set the current working directory inside the container
WORKDIR /app
# Copy go mod and sum files
COPY ./app/go.mod ./app/go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download
# Copy the source from the current directory to the working Directory inside the container
COPY ./app .
# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
# Run stage
FROM alpine:3.17
# Set the current working directory inside the container
WORKDIR /app
# Copy the Pre-built binary file from the previous stage
COPY --from=build-env /app/main .
COPY ./config-files/prod.json ./config.json
# Expose port 8080 to the outside world
EXPOSE 8080
# Command to run the executable
CMD ["./main"]