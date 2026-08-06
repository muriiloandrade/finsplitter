# Build stage
FROM golang:1.26.5-trixie@sha256:87ffdb09b6a2e29ff910748b745395e8a0299aa80b7c0551cdca9b55e3fd2b3e AS setup

WORKDIR /app

ENV GOEXPERIMENT=jsonv2

COPY go.* .

RUN go mod download

COPY . .

# Build the application
FROM setup AS builder

ARG GIT_COMMIT
ARG GIT_BUILD_TAG
ARG BUILD_TIME

RUN GOEXPERIMENT=jsonv2 CGO_ENABLED=0 go build \ 
    -ldflags "-X main.BuildCommit=${GIT_COMMIT} -X main.BuildTime=${BUILD_TIME} -X main.BuildTag=${GIT_BUILD_TAG}" \
    -o bin/finsplitter cmd/api/main.go

# Execution stage
FROM gcr.io/distroless/static-debian12:nonroot@sha256:f5b485ea962d9bd1186b2f6b3a061191539b905b82ec395de78cbfae51f20e35 AS production

# Copy the built binary
COPY --from=builder /app/bin/finsplitter /

# Execute the application
CMD ["/finsplitter"]
