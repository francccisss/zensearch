FROM golang:alpine

WORKDIR /app

COPY . .

RUN go mod tidy


RUN apk update && \
    apk add --no-cache \
    xvfb \
    chromium \
    xauth


RUN go build

CMD ["./web-crawler-service"]
