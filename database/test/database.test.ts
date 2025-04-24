import test from "node:test";
import path from "node:path";
import { execScripts, initDatabase } from "./setup_db.ts";
import database from "../src/database.ts";
import type { IndexedWebpage, URLs } from "../src/utils/types.ts";

const wc = path.join(import.meta.dirname, "./website_collection_tst.db");
const fq = path.join(import.meta.dirname, "./frontier_queue_tst.db");
const websitesDB = initDatabase(wc);
const frontierQueueDB = initDatabase(fq);
execScripts(websitesDB, path.join(import.meta.dirname, "./init.sql"));
execScripts(frontierQueueDB, path.join(import.meta.dirname, "./front.sql"));

export const testPages: IndexedWebpage[] = [
  {
    Message: "Crawled successfully",
    CrawlStatus: 200,
    Webpage: {
      Header: {
        Title: "Example Page 1",
        Url: "https://lol.com/page1",
      },
      Contents:
        "<html><head><title>Example Page 1</title></head><body>Hello, world!</body></html>",
    },
    URLSeed: "https://lol.com/",
  },
  {
    Message: "Crawled successfully",
    CrawlStatus: 200,
    Webpage: {
      Header: {
        Title: "Example Page 2",
        Url: "https://rerere.com/page2",
      },
      Contents:
        "<html><head><title>Example Page 2</title></head><body>Test content here.</body></html>",
    },
    URLSeed: "https://rerere.com/",
  },
  {
    Message: "Page not found",
    CrawlStatus: 404,
    Webpage: {
      Header: {
        Title: "Not Found",
        Url: "https://unsolicitedadvice.com/",
      },
      Contents: "",
    },
    URLSeed: "https://unsolicitedadvice.com/",
  },
  {
    Message: "Redirected",
    CrawlStatus: 301,
    Webpage: {
      Header: {
        Title: "Redirect Page",
        Url: "https://excel.com/redirect",
      },
      Contents: "",
    },
    URLSeed: "https://excel.com/redirect",
  },
  {
    Message: "Crawled successfully",
    CrawlStatus: 200,
    Webpage: {
      Header: {
        Title: "API Response",
        Url: "https://api.example.com/data",
      },
      Contents: '{"message": "API response data"}',
    },
    URLSeed: "https://api.example.com/data",
  },
];

test.test("Webpage Indexing", async (t) => {
  for (let page of testPages) {
    try {
      await database.saveWebpage(websitesDB, page);
    } catch (e) {
      console.error(e);
      if (e.code.toLowerCase().includes("unique")) {
        console.error("TUPLE ALREADY EXISTS");
      }
      t.assert.fail("Not able to save indexed webpage");
    }
  }
});

test.test("clear_websitesDB", async () => {
  websitesDB.prepare("delete from webpages;").run();
  websitesDB.prepare("delete from indexed_sites;").run();
  console.log("cleared database");
});

test.test("Webpage Query", async (t) => {
  try {
    const webpages = database.queryWebpages(websitesDB);
    console.log(webpages);
  } catch (e) {
    t.assert.fail(e.code);
  }
});

//test.test("Check existing websites", (t) => {
//  try {
//    const crawlList = ["https://api.example.com"];
//    const webpages = database.checkAlreadyIndexedWebpage(websitesDB, crawlList);
//    console.log("Existing Urls: ", webpages);
//  } catch (e) {
//    console.error(e);
//    t.assert.fail(e.code);
//  }
//});
//
test.test("dequeue", (t) => {
  const urls: URLs = {
    Domain: "https://example.com",
    Nodes: [
      "https://example.com/about",
      "https://example.com/contact",
      "https://example.com/blog",
      "https://example.com/products/item-1",
      "https://example.com/products/item-2",
    ],
  };
  try {
    database.enqueueUrls(frontierQueueDB, urls);
    const node = database.dequeueURL(frontierQueueDB, urls.Domain);
    console.log("DEQUEUED NODE: ", node);
  } catch (e) {
    console.error(e);
    frontierQueueDB.prepare("DELETE FROM nodes").run();
    t.assert.fail(e.code);
  } finally {
    console.log(frontierQueueDB.prepare("select * from nodes").all());
    frontierQueueDB.prepare("DELETE FROM nodes").run();
  }
});

test.test("enqueue urls", (t) => {
  const urls: URLs = {
    Domain: "https://example.com",
    Nodes: [
      "https://example.com/about",
      "https://example.com/contact",
      "https://example.com/blog",
      "https://example.com/products/item-1",
      "https://example.com/products/item-2",
    ],
  };
  try {
    database.enqueueUrls(frontierQueueDB, urls);
  } catch (e) {
    console.error(e);
    t.assert.fail(e.code);
  } finally {
    console.log(frontierQueueDB.prepare("select * from nodes").all());
    console.log(frontierQueueDB.prepare("select * from queues").all());
  }
});

test.only("Dequeuing and Enqueuing Urls", (t) => {
  let urls: URLs = {
    Domain: "https://example.com",
    Nodes: [
      "https://example.com/about",
      "https://example.com/contact",
      "https://example.com/blog",
      "https://example.com/products/item-1",
      "https://example.com/products/item-2",
      "https://example.com/products/item-1",
      "https://example.com/products/item-2",
    ],
  };

  try {
    frontierQueueDB.transaction(() => {
      database.enqueueUrls(frontierQueueDB, urls);

      urls.Nodes.forEach((n) => {
        const dequeued = database.dequeueURL(frontierQueueDB, urls.Domain);
        if (dequeued.inProgressNode == null) {
          return;
        }
        database.setNodeToVisited(frontierQueueDB, dequeued.inProgressNode);
        console.log("DEQUEUD NODE: ", dequeued);
      });
    })();

    urls = {
      Domain: "https://example.com",
      Nodes: [
        "https://example.com/about",
        "https://example.com/contact",
        "https://example.com/blog",
      ],
    };
    //const node = database.setNodeToVisited(frontierQueueDB, "1");
  } catch (e) {
    console.error(e);
    t.assert.fail(e.code);
  } finally {
    console.log(
      "NODES: ",
      frontierQueueDB.prepare("select * FROM nodes").all(),
    );
    console.log(
      "V_NODES: ",
      frontierQueueDB.prepare("select * FROM visited_nodes").all(),
    );
    console.log(
      "QUEUES: ",
      frontierQueueDB.prepare("select * FROM queues").all(),
    );
    frontierQueueDB.prepare("DELETE FROM visited_nodes").run();
    frontierQueueDB.prepare("DELETE FROM nodes").run();
    frontierQueueDB.prepare("DELETE FROM queues").run();
  }
});
