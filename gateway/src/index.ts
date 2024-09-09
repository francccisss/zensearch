import express, { Request, Response, NextFunction } from "express";
import path from "path";
import amqp, { Connection, Channel } from "amqplib";
import connect_rabbitmq from "./rabbitmq";
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
    }
    next();
  },
);

app.get("/", (req: Request, res: Response) => {
  res.sendFile(path.join(__dirname, "public", "index.html"));
});

app.post("/crawl", async (req: Request, res: Response, next: NextFunction) => {
  const docs = [
    "https://fzaid.vercel.app/",
    "https://docs.python.o3/",
    "https://developer.mozilla.org/en-US/docs/Web/JavaScript",
    "https://go.dev/doc/",
    "https://motherfuckingwebsite.com/",
  ];
  const encoder = new TextEncoder();
  const encoded_docs = encoder.encode(JSON.stringify({ docs }));

  console.log("Crawl");
  try {
    const queue = "crawl_rpc_queue";
    const message = "Start Crawl";
    const corID = "f27ac58-7bee-4e90-93ef-8bc08a37e26c";
    const connection = await connect_rabbitmq();
    if (connection === null) throw new Error("TCP Connection lost.");
    const channel = await connection.createChannel();
    const response_queue = await channel.assertQueue("polling_worker_queue", {
      exclusive: false,
    });
    await channel.sendToQueue(queue, Buffer.from(encoded_docs.buffer), {
      replyTo: response_queue.queue,
      correlationId: corID,
    });

    await channel.close();
    res.cookie("job_id", corID);
    res.cookie("job_queue", response_queue.queue);
    res.send("<p>Crawling...</p>");
  } catch (err) {
    const error = err as Error;
    console.log("LOG:Something went wrong with RPC queue");
    console.error(error.message);
    next(err);
  }
});

app.get("/job", async (req: Request, res: Response, next: NextFunction) => {
  const { job_id, job_queue } = req.query;
  if (job_id === undefined || job_queue === undefined)
    throw new Error("There's no worker queue for this session for crawling.");

  try {
    const connection = await connect_rabbitmq();
    if (connection === null) throw new Error("TCP Connection lost.");
    const channel = await connection.createChannel();
    const { queue, messageCount, consumerCount } = await channel.checkQueue(
      job_queue as string,
    );
    if (messageCount === 0) {
      res.send("Crawling...");
    } else {
      const consumer = await channel.consume(
        job_queue as string,
        async (response) => {
          if (response === null) throw new Error("No Response");
          if (response.properties.correlationId === job_id) {
            console.log(
              "LOG: Response from crawler received: %s",
              response.content.toString(),
            );
            console.log("CONSUMED");
            res.clearCookie("job_id");
            res.clearCookie("job_queue");
            res.send("Success");
            await channel.close();
          }
        },
        { noAck: true },
      );
    }
  } catch (err) {
    const error = err as Error;
    console.log("LOG:Something went wrong with RPC queue");
    console.error(error.message);
    next(err);
  }
});

app.post("/search", async (req: Request, res: Response, next: NextFunction) => {
  try {
    const connection = await connect_rabbitmq();
    if (connection === null) throw new Error("TCP Connection lost.");
    const { search } = req.body;
    const channel = await connection.createChannel();
    const queue = "search_queue";
    const rps_queue = "search_rps_queue";
    const cor_id = "a29a5dec-fd24-4db4-83f1-db6dbefdaa6b";
    await channel.assertQueue(queue, {
      exclusive: false,
      durable: false,
    });
    await channel.sendToQueue(queue, Buffer.from(search));
    console.log(search);
    res.send("<p>Results</p>");
    await channel.close();
  } catch (err) {
    const error = err as Error;
    console.error(error.message);
    next(err);
  }
});

app.get("/search", (req: Request, res: Response) => {
  res.sendFile(path.join(__dirname, "public", "search.html"));
});
app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  console.error(err.stack);
  res.status(500).send("Something went wrong!");
});

// User sends search query -> search_query message broker -> received by search engine service
// send database_query to message broker -> database__search_query -> database parses message and queries
// webpages from sqlite, callback queue sends back a new message to a new message queue -> database_response_queue
// -> received by the search engine service, parse the webpage dataset to an array of webpages struct, process
// the webpages by calculating tf-idf and rank them, pass message back to gateway service -> publish_queue -> ??
//
// How can the gateway service received the message from the message queue publish_queue?
// since the gateway service is an express app and is obviosuly a server that would only take in Request
// and not push any data to the client unless the client request's something from the server.
//
// Websocket? but the client needs to upgrade to a websocket connection
