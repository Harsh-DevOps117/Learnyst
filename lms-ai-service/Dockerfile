FROM golang:1.26.3-alpine

WORKDIR /app

# Ca-certificates are required to make HTTPS requests to the Groq API
RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/api

EXPOSE 3002

CMD ["./main"]
