import express, { Request, Response, NextFunction } from "express";
import amqp, { Connection, Channel, ConsumeMessage } from "amqplib";
import path from "path";
import { v4 as uuidv4 } from "uuid";
import {
  CRAWL_QUEUE_CB,
  CRAWL_QUEUE,
  SEARCH_QUEUE_CB,
} from "../rabbitmq/routing_keys";
import rabbitmq from "../rabbitmq";
rabbitmq;

const cors = require("cors");
const body_parser = require("body-parser");
const app = express();
const public_route = [__dirname, "..", "public"];

app.use(body_parser.urlencoded({ extended: false }));
app.use(body_parser.json());
app.use(cors());
app.use(express.static(path.join(...public_route)));
app.use(
  express.static(path.join(__dirname, "..", "public/scripts")),
  (req: Request, res: Response, next: NextFunction) => {
    if (req.path.endsWith(".js")) {
      res.setHeader("Content-Type", "application/javascript");
    }
    next();
  },
);
app.get("/", (req: Request, res: Response) => {
  res.sendFile(path.join(...public_route, "index.html"));
});

// TODO use Websockets for crawling instead of polling like a biiiitchh
app.post("/crawl", async (req: Request, res: Response, next: NextFunction) => {
  const Docs = [
    "https://fzaid.vercel.app/",
    "https://robbowen.digital/",
    "https://naren200.github.io/",
    "https://brittanychiang.com",
  ];

  const encoder = new TextEncoder();
  const encoded_docs = encoder.encode(JSON.stringify({ Docs }));

  console.log("Crawl");
  const job_id = uuidv4();
  try {
    const connection = await rabbitmq.connect();
    if (connection === null)
      throw new Error("Unable to create a channel for crawl queue.");
    const channel = await connection.createChannel();

    /*
      sends a message to the database service to check and see if the DOCS
      or list of websites the users want to crawl already exists in the database.
    */

    const db_check_queue = "db_check_express";
    const db_check_response_queue = "db_cbq_express";
    await channel.assertQueue(db_check_queue, {
      durable: false,
      exclusive: false,
    });
    await channel.assertQueue(db_check_response_queue, {
      durable: false,
      exclusive: false,
    });

    // but we can consume it here after checking the database if it exists or not
    // right after it returns with a value.
    await channel.sendToQueue(db_check_queue, Buffer.from(encoded_docs));
    await channel.consume(db_check_response_queue, async (data) => {
      if (data === null) {
        throw new Error("ERROR: Data received is null.");
      }
      try {
        /*
         check if it is empty or not, if so send back a message to the client
         that the list provided have already been crawled
        */
        const undindexed_sites: { Docs: Array<string> } = JSON.parse(
          data.content.toString(),
        );
        if (undindexed_sites.Docs.length === 0) {
          console.log("This shits empty YEEET!");
          res.send("<p>List already crawled.</p>");
        }

        // if not then proceed to crawler service
        const success = await rabbitmq.crawl_job(channel, data.content, {
          queue: CRAWL_QUEUE,
          id: job_id,
        });
        if (!success) {
          next("an Error occured while starting the crawl.");
        }
        channel.ack(data);
      } catch (err) {
        channel.nack(data, false, false);
      }
    });

    /*
      Clients does not need to know if it exists or not, we can still handle it internally.
      We can just send back to the client an ok response after sending a crawl task
      to the crawler service.
    */

    /*
      Creates a session cookie for job polling using the poll route handler `/job`
      the CRAWL_QUEUE_CB is used to poll the crawler service to check and see if
      crawling is done or not you can read the code with the route `/job`
    */

    res.cookie("job_id", job_id);
    res.cookie("job_queue", CRAWL_QUEUE_CB);
    res.cookie("poll_type", "crawling");
    res.send("<p>Crawling...</p>");
  } catch (err) {
    const error = err as Error;
    console.log("LOG:Something went wrong with Crawl queue");
    console.error(error.message);
    next(err);
  }
});

/*
  A Route handler responsible for polling the crawl task, this polls for results
  from the crawler service in a fixed time, it uses the `job_id` to identify the
  correlationId of the message in the message queue, and a `job_queue` to specify
  which message queue it wants to consume from.
*/

app.get("/job", async (req: Request, res: Response, next: NextFunction) => {
  const { job_id, job_queue } = req.query;
  if (job_id === undefined || job_queue === undefined)
    throw new Error("There's no worker queue for this session for crawling.");

  try {
    const connection = await rabbitmq.connect();
    if (connection === null) throw new Error("TCP Connection lost.");
    const channel = await connection.createChannel();
    const job = await rabbitmq.poll_job(channel, {
      id: job_id as string,
      queue: job_queue as string,
    });
    if (!job.done) {
      res.json({ message: "Processing..." });
      return;
    }
    channel.close();
    res.clearCookie("job_id");
    res.clearCookie("job_queue");
    res.clearCookie("poll_type");
    res.json({ message: job.data }).status(200);
  } catch (err) {
    const error = err as Error;
    console.log("LOG:Something went wrong with polling queue");
    console.error(error.message);
    next(err);
  }
});

/*
  Hmm might change this next time idk.

  Upgrades http protocol to websocket so that we dont need to poll
  for the search engine service's search results, and instead we can just create an
  persistent tcp connection using websocket protocol, the websocket server will
  be responsible for listening/consuming the search results from the
  search engine service, and then send back the search result data back to the client.

  - search queue handlers for rabbitmq is in the `rabbitmq/` folder
  - search channel listeners are in the `websocket/` folder

*/

app.get("/search", async (req: Request, res: Response, next: NextFunction) => {
  const job_id = uuidv4();
  try {
    res.setHeader("Connection", "Upgrade");
    res.setHeader("Upgrade", "Websocket");

    /*
      Need to job_id such that different messages in the message queue `SEARCH_QUEUE_CB`,
      the websocket listener will be able to determine which job's which.
      eg: user sends "fzaid projects" search query then that will have its own job_id
      specifically for that search query in the message queue.
    */

    res.cookie("job_id", job_id);
    res.sendFile(path.join(...public_route, "search.html"));
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

export default app;
