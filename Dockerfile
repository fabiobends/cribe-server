FROM golang:1.23-alpine

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o cribe-server ./cmd/app

EXPOSE 8080

CMD ["./cribe-server"]
