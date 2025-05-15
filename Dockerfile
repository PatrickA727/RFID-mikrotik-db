# ---- Stage 1: Build ----
    FROM golang:1.22 AS builder

    # Set the working directory for building stage (All source files and dependencies)
    WORKDIR /app
    
    # Copy Go modules files first for better caching
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy the rest of the app
    COPY . .
    
    # Build the Go app (output binary named 'app')
    RUN go build -o app .
    
    # Uses distroless(lightweight) debian as the os for the image
    FROM gcr.io/distroless/base-debian12 AS runner
    
    # Set working directory for runtime in the debian os
    WORKDIR /root/

    COPY --from=builder /app/.env .
    
    # Copy the compiled binary from the builder stage
    COPY --from=builder /app/app .
    
    # Expose the application port
    EXPOSE 5000
    
    # Run the binary
    CMD ["./app"]
    