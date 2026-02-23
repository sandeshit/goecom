# -------- Builder --------
FROM golang:1.25-bookworm AS builder

# Install git (needed for go modules)
RUN apt-get update && \
    apt-get install -y --no-install-suggests --no-install-recommends git && \
    apt-get clean

WORKDIR /code

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all code
COPY . .

# Build binary explicitly for Linux, statically
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /code/app ./cmd

# -------- Runtime --------
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /code

# Copy binary from builder
COPY --from=builder /code/app /code/app

EXPOSE 8080

# Run binary using absolute path
CMD ["/code/app"]
