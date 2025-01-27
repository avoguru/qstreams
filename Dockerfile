# Use the official Go image as the base image
FROM golang:1.23.5

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project into the container
COPY . .

# Expose the server port
EXPOSE 8080

# Command to run the Go application
CMD ["go", "run", "./server/main.go"]