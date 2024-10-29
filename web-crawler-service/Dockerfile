FROM golang:alpine

WORKDIR /app

COPY . .

RUN go mod tidy

RUN apk update && \
    apk add --no-cache \
    xvfb \
    chromium \
    xauth \
    chromium-chromedriver

ENV PATH="/usr/bin:/usr/local/bin:${PATH}"

RUN go build

CMD ["./web-crawler-service"]
