# ---- Stage 1: Builder --------------------------------------------------------

FROM golang:1.26-alpine AS builder

WORKDIR /app

# git is needed for go mod download with some dependencies
RUN apk add --no-cache git

# Download dependencies first — cached as a separate layer.
# Only re-runs when go.mod or go.sum changes, not on every code change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0  — pure Go binary, no C dependencies
# -ldflags       — strip debug info and symbol table → smaller binary
# ./cmd/server/. — compile entire package, not just main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w" \
    -o server \
    ./cmd/server/...

# ---- Stage 2: Runtime --------------------------------------------------------
FROM alpine:3.20

WORKDIR /app

# ca-certificates — needed for outbound HTTPS calls (Mailjet etc.)
# tzdata         — needed for correct timezone handling in logs and DB
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user to run the binary — never run as root in production
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy only the compiled binary from builder — nothing else
COPY --from=builder /app/server .

# Switch to non-root user
USER appuser

EXPOSE 7003

CMD ["./server"]