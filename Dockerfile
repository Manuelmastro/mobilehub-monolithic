# Stage 1: Build stage
FROM golang:1.22-alpine AS builder

# Install required dependencies for building
RUN apk add --no-cache dos2unix postgresql-client

# Set the working directory
WORKDIR /app

# Copy Go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Convert line endings to Unix-style and make script executable
RUN dos2unix wait-for-postgres.sh && chmod +x wait-for-postgres.sh

# Build the application
RUN go build -o main .

# Stage 2: Runtime stage
FROM alpine:latest

# Install PostgreSQL client for runtime if needed
RUN apk add --no-cache postgresql-client

# Copy the necessary files from the builder stage
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/wait-for-postgres.sh /usr/local/bin/wait-for-postgres.sh

# Ensure the script is executable
RUN chmod +x /usr/local/bin/wait-for-postgres.sh

# Expose the port your app runs on
EXPOSE 8080

# Command to run the application
CMD ["wait-for-postgres.sh", "./main"]
