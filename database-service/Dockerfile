FROM node:20.16.0

WORKDIR /app

COPY . .

RUN npm install
RUN npm run build

EXPOSE 8080

CMD ["node", "./dist/index.js"]
