import express from "express";
import type { Request, Response, NextFunction } from "express";
import path from "path";
import { v4 as uuidv4 } from "uuid";
import rabbitmq from "../rabbitmq/index.js";
import { create } from "express-handlebars";
import segmentSerializer from "../segments/segment_serializer.js";
import { EXPRESS_CRAWLER_QUEUE } from "../rabbitmq/routing_keys.js";

import cors from "cors";
import body_parser from "body-parser";
const app = express();
const public_route = [import.meta.dirname, "../", "public"];

app.use(body_parser.urlencoded({ extended: false }));
app.use(body_parser.json());
app.use(cors());

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
app.set("views", path.join(import.meta.dirname, "views"));
// app.use(
//   express.static(path.join(...public_route)),
// (req: Request, res: Response, next: NextFunction) => {
//   if (req.path.endsWith(".js")) {
//     res.setHeader("Content-Type", "application/javascript");
//     next();
//   }
//   if (req.path.endsWith(".css")) {
//     res.setHeader("Content-Type", "text/css");
//     next();
//   }
// },
// );
app.use(express.static(path.join(...public_route)));
app.get("/", (req: Request, res: Response) => {
  console.log("NEW CONNECTION");
  console.log(req.cookies);
  res.sendFile(path.join(...public_route, "index.html"));
});

app.post("/crawl", async (req: Request, res: Response, next: NextFunction) => {
  // check if URL is valid
  const Docs: Array<string> = [...req.body];
  if (Docs.length == 0) {
    return res.status(200).json({
      is_crawling: false,
      message: "List is empty.",
      crawl_list: Docs,
    });
  }
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
    // results are the array of websites that have NOT been indexed yet

    const results = await rabbitmq.client.crawlListCheck(encoded_docs);
    console.log(results);
    // let results = { unindexed: Docs };
    if (results === null) {
      throw new Error("Unable to check user's crawl list, results == null.");
    }
    if (results.unindexed.length === 0) {
      return res.status(200).json({
        is_crawling: false,
        message:
          "All of the items in this list has already indexed, please provide a new list.",
        crawl_list: Docs,
      });
    }

    /*
      Need to notify users that some of the items in the list have already been indexed,
      so we need to send back the items that were NOT included in the unindexed list
      (means return only the indexed ones).

      Doing the opposite by filtering out websites that have already been indexed from `Docs` and
      return it back to the user to change these entries.

      This returns the unindexed list to the user
    */
    if (results.unindexed.length !== Docs.length) {
      return res.status(200).json({
        is_crawling: false,
        message: "Some of the items in this list has already been indexed.",
        crawl_list: Docs.filter(
          (website) => results.unindexed.includes(website) ?? website,
        ),
      });
    }

    // proceed to Crawler Service

    res.cookie("job_id", job_id);
    res.cookie("job_count", results.unindexed.length);
    res.cookie("job_queue", EXPRESS_CRAWLER_QUEUE);
    res.cookie("message_type", "crawling");
    res.setHeader("Connection", "Upgrade");
    res.setHeader("Upgrade", "Websocket");
    res.json({
      is_crawling: true,
      message: "Crawling",
      crawl_list: results.unindexed,
    });
  } catch (err) {
    next(err);
  }
});

/*
  Hmm might change this next time idk.

  As of now, im creating and desstroying a listener for every requests
  this might add some significant overhead that im not aware of but right now
  it works, i could change this by using an event emmiter and just call the listener
  on init.


  Problem not receiving all of the segments from search engine.
*/

app.get("/search", async (req: Request, res: Response, next: NextFunction) => {
  const q = req.query.q;

  console.log("NOTIF: Search Query sent");
  console.log("SEARCH QUERY: %s", q);
  try {
    const isSent = await rabbitmq.client.sendSearchQuery(q as string);
    if (!isSent) {
      throw new Error("ERROR: Unable to send search query.");
    }

    const webpageBuffer = await segmentSerializer.listenIncomingSegments(
      rabbitmq.client.searchChannel!,
      rabbitmq.client.segmentGenerator.bind(rabbitmq.client),
    );
    rabbitmq.client.eventEmitter.emit("done", {});
    const parseWebpages = segmentSerializer.parseWebpages(webpageBuffer);
    // .slice(0, 10);

    res.render("search", {
      search_results: parseWebpages.length === 0 ? [] : parseWebpages,
      query: q,
    });
  } catch (err) {
    const error = err as Error;
    console.error(error.message);
    next(err);
  }
});

app.use((err: Error, _: Request, res: Response, __: NextFunction) => {
  console.error("ERROR STACK: %s", err.stack);
  res.status(500).json({ message: err.message });
});

export default app;
