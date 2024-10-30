FROM golang:alpine

WORKDIR /app

COPY . .


RUN apk update

ENV PATH="/usr/bin:/usr/local/bin:${PATH}"

RUN go mod tidy

RUN go build

ENTRYPOINT ["go","test","-args"]

# Handshake response does not match any supported protocol.
#Response payload: {"sessionId":"0afb1b1c3ff9da91c061209f6a75a9f3","status":33,"value":{"message":"session not created: Missing or invalid capabilities
#(Driver info: chromedriver=130.0.6723.69 (3ec172b971b9478c515a58f591112c5c23fa4965-refs/branch-heads/6723@{#1452}),platform=Linux 6.10.13-3-MANJARO x86_64)"}}
CMD ["https://fzaid.vercel.app/","https://gobyexample.com"]
