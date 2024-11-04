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
import { create } from "express-handlebars";

const cors = require("cors");
const body_parser = require("body-parser");
const app = express();
const public_route = [__dirname, "..", "public"];

app.use(body_parser.urlencoded({ extended: false }));
app.use(body_parser.json());
app.use(cors());
app.use(express.static(path.join(...public_route)));

app.engine(
  "handlebars",
  create({
    runtimeOptions: {
      allowProtoPropertiesByDefault: true,
    },

    helpers: {
      checkLength: (text: string): string => {
        return text.trim().length === 0 || text === undefined
          ? "No description."
          : text;
      },
      urlOrigin: (url: string): string => {
        const u = new URL(url).hostname;
        const l = u.split(".");
        // example.com -> [example,com] length=2
        // docs.example.com -> [docs,example,com] length=3
        return l.length > 2 ? l[1] : l[0];
      },
      textInitial: (title: string): string => {
        return title[0];
      },
      crumbs: (url: string): string => {
        const u = new URL(url);
        const path = u.pathname;
        return path !== "/"
          ? u.hostname + path.split("/").join(" > ")
          : u.hostname;
      },

      noResults: (results: []): boolean => {
        return results.length == 0;
      },
    },
  }).engine,
);
app.set("view engine", "handlebars");
app.set("views", path.join(__dirname, "views"));
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
    console.log("ERROR :Something went wrong with Crawl queue");
    console.error(error.message);
    next(err);
  }
});

/*
  Hmm might change this next time idk.

  As of now, im creating and desstroying a listener for every requests
  this might add some significant overhead that im not aware of but right now
  it works, i could change this by using an event emmiter and just call the listener
  on init.
*/

app.get("/search", async (req: Request, res: Response, next: NextFunction) => {
  const q = req.query.q;
  function eventListener(msg: { data: ConsumeMessage; err: Error | null }) {
    if (msg.err !== null) {
      throw new Error(msg.err.message);
    }
    const parse_ranked_pages = JSON.parse(msg.data.content.toString());
    console.log("NOTIF: Search query sent to the client .");

    res.render("search", {
      search_results: parse_ranked_pages.length === 0 ? [] : parse_ranked_pages,
      query: q,
    });
    //res.json({ msg: parse_ranked_pages, success: true, query: q });
    rabbitmq.client.eventEmitter.removeAllListeners("searchResults");
  }
  try {
    console.log("NOTIF: Search Query sent");
    console.log("SEARCH QUERY: %s", q);
    // Creates a search channel
    const is_sent = await rabbitmq.client.send_search_query(q as string);
    if (!is_sent) {
      throw new Error("ERROR: Unable to send search query.");
    }
    rabbitmq.client.eventEmitter.on("searchResults", eventListener);
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
