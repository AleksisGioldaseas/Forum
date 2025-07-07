# Build stage
FROM golang:1.24-alpine AS builder

# Install dependencies
RUN apk add --no-cache sqlite openssl musl-dev gcc make

WORKDIR /app
COPY . .

# Generate certs and run/populate database

RUN go build -o /app/forum-app app/main.go

# Runtime stage
FROM alpine:latest

# Runtime dependencies
RUN apk add --no-cache sqlite ca-certificates bash tzdata

COPY --from=builder /app/forum-app /forum/forum-app
COPY --from=builder /app/sql /forum/sql
COPY --from=builder /app/web /forum/web
COPY --from=builder /app/populate_data /forum/populate_data

WORKDIR /forum

EXPOSE 8080

# Run first time setup
COPY entry.sh /forum/
RUN chmod +x /forum/entry.sh
ENTRYPOINT ["/forum/entry.sh"]
CMD ["/forum/forum-app"]