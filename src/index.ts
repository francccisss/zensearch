import express, { Request, Response, NextFunction } from "express";
import path from "path";
import amqp, { Connection, Channel } from "amqplib";
const cors = require("cors");
const body_parser = require("body-parser");
const app = express();
const PORT = 8080;

app.listen(PORT, () => {
  console.log("Listening to Port:", PORT);
});

app.use(body_parser.urlencoded({ extended: false }));
app.use(body_parser.json());
app.use(cors());
app.use(express.static(path.join(__dirname, "public")));
app.use(
  express.static(path.join(__dirname, "public/scripts")),
  (req: Request, res: Response, next: NextFunction) => {
    if (req.path.endsWith(".js")) {
      res.setHeader("Content-Type", "application/javascript");
      next();
    }
    next();
  },
);

app.get("/", (req: Request, res: Response) => {
  res.sendFile(path.join(__dirname, "public", "index.html"));
});

app.post("/crawl", async (req: Request, res: Response) => {
  console.log("Crawl");
  const connection = await amqp.connect("amqp://localhost");
  const channel = await connection.createChannel();
  const queue = "crawl_rpc_queue";
  const message = "Start Crawl";
  const corID = "f27ac58-7bee-4e90-93ef-8bc08a37e26c";
  const response_queue = await channel.assertQueue("crawl_reply_queue", {
    exclusive: true,
  });
  channel.sendToQueue(queue, Buffer.from(message), {
    replyTo: response_queue.queue,
    correlationId: corID,
  });
  channel.consume(
    response_queue.queue,
    async (msg) => {
      if (msg === null) throw new Error("No Message");
      console.log(msg);
      if (msg.properties.correlationId == corID) {
        console.log(" [.] Got %s", msg.content.toString());
        console.log("[x] Sent %s", message);
        res.send("<p>Crawling...</p>");
      }
    },

    {
      noAck: true,
    },
  );
});

app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  console.error(err.stack);
  res.status(500).send("Something went wrong!");
});
