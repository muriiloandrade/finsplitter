# Build stage
FROM golang:1.24.1 AS setup

WORKDIR /app

COPY go.* .

RUN go mod download

COPY . .

FROM setup AS builder

RUN CGO_ENABLED=0 go build -o bin/finsplitter cmd/api/main.go

# Execution stage
FROM gcr.io/distroless/base-debian10 AS production

# Copy the built binary
COPY --from=builder /app/bin/finsplitter /

# Execute the application
CMD ["/finsplitter"]
