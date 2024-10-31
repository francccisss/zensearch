FROM golang:alpine

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build

CMD ["./search-engine"]
