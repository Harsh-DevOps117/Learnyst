FROM golang:1.26.3-alpine

# Install ffmpeg for adaptive streaming
RUN apk add --no-cache ffmpeg

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/worker

CMD ["./main"]
