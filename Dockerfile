# Use the official Go image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source code
COPY . ./

# Build the Go application
RUN go build -o main .

# Expose the port your application listens on
EXPOSE 8080 

# Define the command to run your application
CMD ["./main"]
