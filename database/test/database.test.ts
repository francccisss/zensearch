import test from "node:test";
import path from "node:path";
import { execScripts, initDatabase } from "./setup_db.ts";
import database from "../src/database.ts";
import type { IndexedWebpage } from "../src/utils/types.ts";

const wc = path.join(import.meta.dirname, "./db.test.db");
const db = initDatabase(wc);
execScripts(db, path.join(import.meta.dirname, "./init.sql"));

export const testPages: IndexedWebpage[] = [
  {
    Message: "Crawled successfully",
    CrawlStatus: 200,
    Webpage: {
      Header: {
        Title: "Example Page 1",
        Url: "https://example.com/page1",
      },
      Contents:
        "<html><head><title>Example Page 1</title></head><body>Hello, world!</body></html>",
    },
    URLSeed: "https://example.com/page1",
  },
  {
    Message: "Crawled successfully",
    CrawlStatus: 200,
    Webpage: {
      Header: {
        Title: "Example Page 2",
        Url: "https://example.com/page2",
      },
      Contents:
        "<html><head><title>Example Page 2</title></head><body>Test content here.</body></html>",
    },
    URLSeed: "https://example.com/page2",
  },
  {
    Message: "Page not found",
    CrawlStatus: 404,
    Webpage: {
      Header: {
        Title: "Not Found",
        Url: "https://example.com/missing",
      },
      Contents: "",
    },
    URLSeed: "https://example.com/missing",
  },
  {
    Message: "Redirected",
    CrawlStatus: 301,
    Webpage: {
      Header: {
        Title: "Redirect Page",
        Url: "https://example.com/redirect",
      },
      Contents: "",
    },
    URLSeed: "https://example.com/redirect",
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
      await database.saveWebpage(db, page);
    } catch (e) {
      console.error(e);
      if (e.code.toLowerCase().includes("unique")) {
        console.error("TUPLE ALREADY EXISTS");
      }
      t.assert.fail("Not able to save indexed webpage");
    }
  }
});

test.test("Webpage Query", async (t) => {
  try {
    const webpages = database.queryWebpages(db);
    console.log(webpages);
  } catch (e) {
    t.assert.fail(e.code);
  }
});
