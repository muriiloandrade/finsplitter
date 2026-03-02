# Build stage
FROM golang:1.26.0-trixie@sha256:100774df3fd947fbd1f2b674e71215af6bf087b3676d9fcfdf1a9d0ec9e5cb9c AS setup

WORKDIR /app

COPY go.* .

RUN go mod download

COPY . .

# Build the application
FROM setup AS builder

ARG GIT_COMMIT
ARG GIT_BUILD_TAG
ARG BUILD_TIME

RUN CGO_ENABLED=0 go build \ 
    -ldflags "-X main.BuildCommit=${GIT_COMMIT} -X main.BuildTime=${BUILD_TIME} -X main.BuildTag=${GIT_BUILD_TAG}" \
    -o bin/finsplitter cmd/api/main.go

# Execution stage
FROM gcr.io/distroless/static-debian12:nonroot@sha256:a9329520abc449e3b14d5bc3a6ffae065bdde0f02667fa10880c49b35c109fd1 AS production

# Copy the built binary
COPY --from=builder /app/bin/finsplitter /

# Execute the application
CMD ["/finsplitter"]
