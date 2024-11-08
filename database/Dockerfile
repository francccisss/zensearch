FROM node:20.16.0

WORKDIR /app

COPY . .

COPY ./website_collection.db /app/data/

# for editing file within the website collection database
RUN apt update -y && apt upgrade -y && apt install -y sqlite3

RUN npm install
RUN npm run build

EXPOSE 8080

CMD ["node", "./dist/index.js"]
