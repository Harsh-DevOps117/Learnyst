FROM golang:1.26.3-alpine

# Install Node.js, npm, and zip utilities for custom React builds
RUN apk add --no-cache nodejs npm zip unzip

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/builder

CMD ["./main"]
