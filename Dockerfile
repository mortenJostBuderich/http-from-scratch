# Set versoion
FROM golang:1.27.1

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code to /app
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /custom-http-server ./cmd/httpserver
EXPOSE 8080

# Run
CMD ["/custom-http-server"]