FROM node:20.16.0

WORKDIR /app

COPY . .

# for editing file within the website collection database
RUN apt update -y && apt upgrade -y
RUN apt install -y sqlite3

RUN npm install
RUN npm run build

EXPOSE 8080

CMD ["node", "./dist/index.js"]
