# Use the official Golang image as the base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the Go source code into the container
COPY . .

# Build the Go application
RUN go build -o main

# Expose the port your Go application will run on
EXPOSE 8989

# Command to run your Go application
CMD ["./main"]
