# Use an official Golang runtime as a parent image
FROM golang:1.21.1 AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace
COPY go.mod go.sum ./

# Fetch the required third-party package
RUN go mod download

COPY *.go ./

# # Build the Go application inside the container
RUN CGO_ENABLED=0 GOOS=linux go build -o ./go-sctp-server

# # Expose port 38412 for SCTP
# EXPOSE 9999

# # Run the Go application when the container starts
# CMD ["/go-sctp-server"]


FROM debian:bullseye-slim

# Set the working directory inside the container
WORKDIR /app

# Copy only the binary from the build stage
COPY --from=build /app/go-sctp-server .

# Expose port 38412 for SCTP
EXPOSE 38412

# Run the Go application when the container starts
CMD ["./go-sctp-server"]