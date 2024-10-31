FROM golang:alpine

WORKDIR /app

COPY . .

RUN apk update

RUN go mod tidy

RUN go build

CMD ["./crawler"]
