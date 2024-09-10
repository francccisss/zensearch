import express, { Request, Response, NextFunction } from "express";
import path from "path";
import amqp, { Connection, Channel } from "amqplib";
import rabbitmq from "./rabbitmq";
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
  const queue = "crawl_rpc_queue";
  const corID = "f27ac58-7bee-4e90-93ef-8bc08a37e26c";
  try {
    const connection = await rabbitmq.connect();
    if (connection === null)
      throw new Error("Unable to create a channel for crawl queue.");
    const channel = await connection.createChannel();
    const success = await rabbitmq.crawl_job(channel, encoded_docs, {
      queue,
      id: corID,
    });
    if (!success) {
      next("an Error occured while starting the crawl.");
    }
    res.cookie("job_id", corID);
    res.cookie("job_queue", queue);
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
    const connection = await rabbitmq.connect();
    if (connection === null) throw new Error("TCP Connection lost.");
    const channel = await connection.createChannel();
    const is_polling = await rabbitmq.poll_job(channel, {
      id: job_id as string,
      queue: job_queue as string,
    });
    if (!is_polling) {
      res.send("Processing...");
      return;
    }
    res.clearCookie("job_id");
    res.clearCookie("job_queue");
    res.send("Success");
  } catch (err) {
    const error = err as Error;
    console.log("LOG:Something went wrong with polling queue");
    console.error(error.message);
    next(err);
  }
});

app.get("/search", async (req: Request, res: Response, next: NextFunction) => {
  try {
    const connection = await rabbitmq.connect();
    if (connection === null) throw new Error("TCP Connection lost.");
    const q = req.body.q ?? req.query.q;
    await rabbitmq.search_job(q, connection);
    console.log(q);
    res.sendFile(path.join(__dirname, "public", "search.html"));
  } catch (err) {
    const error = err as Error;
    console.error(error.message);
    next(err);
  }
});

app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  console.error(err.stack);
  res.status(500).send("Something went wrong!");
});
