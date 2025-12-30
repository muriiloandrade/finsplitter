# Build stage
FROM golang:1.25.5-trixie@sha256:674418a0e262957bb2f5cb55f2519fb77616379ec6d5722f09ffe8fdfae7660e AS setup

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
FROM gcr.io/distroless/static-debian12:nonroot@sha256:2b7c93f6d6648c11f0e80a48558c8f77885eb0445213b8e69a6a0d7c89fc6ae4 AS production

# Copy the built binary
COPY --from=builder /app/bin/finsplitter /

# Execute the application
CMD ["/finsplitter"]
