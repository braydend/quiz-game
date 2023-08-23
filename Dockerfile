# Build stage for the Svelte frontend
FROM node:16.17-alpine AS frontend-builder
WORKDIR /app/web

# Copy the Svelte frontend source code
COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web .
RUN npm run build

# Final stage for the Go web server
FROM golang:1.20-alpine AS go-builder
WORKDIR /app

# Copy the Go server source code
COPY ./go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

# Final image
FROM alpine:3.14
WORKDIR /app

# Copy the built Go server
COPY --from=go-builder /app/server .

# Copy the built Svelte frontend
COPY --from=frontend-builder /app/web/dist ./web/dist

# Expose the desired port
EXPOSE 8000

# Set the command to run the Go web server
CMD ["./server"]