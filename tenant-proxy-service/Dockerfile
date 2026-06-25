FROM golang:1.26.3-alpine

WORKDIR /app

# Install root CA certificates for HTTPS/SSL outbounding
RUN apk --no-cache add ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/proxy

EXPOSE 80 443 8080

CMD ["./main"]
