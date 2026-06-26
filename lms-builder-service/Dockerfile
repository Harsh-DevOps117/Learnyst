FROM golang:1.26.3-alpine

RUN apk add --no-cache nodejs npm zip unzip

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/builder

CMD ["./main"]
