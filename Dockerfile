# Build stage
FROM golang:1.24.2 AS setup

WORKDIR /app

COPY go.* .

RUN go mod download

COPY . .

# Development stage with Delve debugger
FROM setup AS development

RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Build the application
FROM setup AS builder

ARG GIT_COMMIT
ARG GIT_BUILD_TAG
ARG BUILD_TIME

RUN CGO_ENABLED=0 go build \ 
    -ldflags "-X main.BuildCommit=${GIT_COMMIT} -X main.BuildTime=${BUILD_TIME} -X main.BuildTag=${GIT_BUILD_TAG}" \
    -o bin/finsplitter cmd/api/main.go

# Execution stage
FROM gcr.io/distroless/static-debian12:nonroot AS production

# Copy the built binary
COPY --from=builder /app/bin/finsplitter /

# Execute the application
CMD ["/finsplitter"]
