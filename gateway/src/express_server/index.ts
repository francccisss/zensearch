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
import { Data } from "ws";
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
  const Docs: Array<string> = [...req.body];
  const encoder = new TextEncoder();
  const encoded_docs = encoder.encode(JSON.stringify({ Docs }));

  console.log("NOTIF: Crawl request sent.");
  const job_id = uuidv4();
  try {
    /*
     Check the users' crawl list if any of the items in that list exists on the database
     if so then return only the ones that are not indexed.

     If the length of the unindexed list is 0 then we can notify the users to send
     a new list of websites to crawl.

     Since we return the modified list that are unindexed, we then pass it to the
     crawler service through the `crawl(<unindexed_list buffer>)` function
     we then specify that current crawl job with a job_id and a routing key to
     route the messsage.
    */

    // TODO need to return some error after some amount of time if there is not ack received
    const results = await rabbitmq.client.crawl_list_check(encoded_docs);
    if (results === null) {
      throw new Error("Unable to check user's crawl list.");
    }
    if (results.undindexed.length === 0) {
      console.log("This shits empty YEEET!");
      return res.status(200).json({
        is_crawling: false,
        message:
          "All of the items in this list has already indexed, please provide a new list.",
        crawl_list: Docs,
      });
    }

    /*
      Need to notify users that some of the items in the list have already been indexed,
      so we need to send back the items that are not included in the unindexed list
      (means return only the indexed ones).

      Doing the opposite by filtering out websites that have already been indexed and
      return it back to the user to change these entries.
    */
    if (results.undindexed.length !== Docs.length) {
      return res.status(200).json({
        is_crawling: false,
        message: "Some of the items in this list has already been indexed.",
        crawl_list: Docs.filter(
          (website) => !results.undindexed.includes(website) ?? website,
        ),
      });
    }
    console.log({ unindexed: results.undindexed });

    // proceed to Crawler Service

    /*
      Creates a session cookie for job polling using the poll route handler `/job`
      the CRAWL_QUEUE_CB is used to poll the crawler service to check and see if
      crawling is done or not you can read the code with the route `/job`
    */
    res.cookie("job_id", job_id);
    res.cookie("job_count", results.undindexed.length);
    res.cookie("job_queue", CRAWL_QUEUE_CB);
    res.cookie("message_type", "crawling");
    res.setHeader("Connection", "Upgrade");
    res.setHeader("Upgrade", "Websocket");
    res.json({
      is_crawling: true,
      message: "Crawling",
      crawl_list: results.undindexed,
    });
  } catch (err) {
    const error = err as Error;
    console.log("ERRO :Something went wrong with Crawl queue");
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
  const q = req.query.q;
  console.log("NOTIF: Message received from client:  search query: %s", q);
  try {
    const is_sent = await rabbitmq.client.send_search_query(q as string);
    if (!is_sent) {
      throw new Error("ERROR: Unable to send search query.");
    }
    console.log("NOTIF: Search query sent to the search engine service.");

    // Throws an error that will be caught by the route handler.
    const msg = await rabbitmq.client.search_channel_listener();
    // Throws an error that will be caught by the route handler.
    console.log("NOTIF: Results from search engine service retrieved.");

    console.log(msg);
    res.json({ msg, success: true });
  } catch (err) {
    const error = err as Error;
    console.error(error.message);
    next(err);
  }
});

app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  console.error(err.stack);
  res.status(500).json({ message: err.message });
});

export default app;
