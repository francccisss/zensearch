import test from "node:test";
import mysql from "mysql2/promise";
import dbInterface from "../src/database.ts";
import type { IndexedWebpage, URLs } from "../src/utils/types.ts";
import "dotenv/config";
import { exit } from "node:process";
import { readFileSync } from "node:fs";
import path from "node:path";
const poolOption: mysql.PoolOptions = {
  user: process.env.DB_USER,
  password: process.env.DB_PASS,
  database: process.env.DB_NAME,
  host: process.env.DB_HOST,
  multipleStatements: true,
};

async function execScripts(
  db: mysql.Pool | mysql.Connection,
  scriptPath: string,
): Promise<void> {
  console.log(`Executing sql script for ${scriptPath}`);
  if (db === null) {
    console.error("ERROR: database does not exist for %s", scriptPath);
    exit(1);
  }
  const f = readFileSync(scriptPath, "utf-8");
  // TODO: Make sure each table that references a table needs to be triggered right after the
  // referenced table has already been created
  try {
    f.split(";")
      .filter((t) => t.trim())
      .forEach(async (t) => {
        try {
          await db.execute(t);
        } catch (e: any) {
          if (e.message.includes("exists")) {
            console.log("skipping duplicate");
            return;
          }
          console.error(e);
          throw new Error(e);
        }
      });
  } catch (e: any) {
    throw new Error(e);
  }
}

const webpages: IndexedWebpage[] = [
  {
    Message: "Crawl completed successfully",
    CrawlStatus: 200,
    Webpage: {
      Header: {
        Title: "OpenAI",
        URL: "openai.com",
      },
      Contents:
        "OpenAI develops artificial intelligence with the mission of ensuring it benefits humanity.",
    },
    URLSeed: "openai.com",
  },
  {
    Message: "Crawl completed successfully",
    CrawlStatus: 200,
    Webpage: {
      Header: {
        Title: "TypeScript Official Site",
        URL: "www.typescriptlang.org",
      },
      Contents:
        "TypeScript is a strongly typed programming language that builds on JavaScript.",
    },
    URLSeed: "www.typescriptlang.org",
  },
  {
    Message: "Crawl failed: Page not found",
    CrawlStatus: 404,
    Webpage: {
      Header: {
        Title: "Missing Page",
        URL: "somesite.com/404",
      },
      Contents: "The page you are looking for does not exist.",
    },
    URLSeed: "somesite.com/404",
  },
  {
    Message: "Crawl completed successfully",
    CrawlStatus: 200,
    Webpage: {
      Header: {
        Title: "MDN Web Docs",
        URL: "developer.mozilla.org",
      },
      Contents:
        "MDN provides documentation for web developers on HTML, CSS, and JavaScript.",
    },
    URLSeed: "developer.mozilla.org",
  },
  {
    Message: "Crawl failed: Timeout",
    CrawlStatus: 408,
    Webpage: {
      Header: {
        Title: "Example Domain",
        URL: "example.com",
      },
      Contents: "",
    },
    URLSeed: "example.com",
  },
];

test.suite("Webpage indexing", async () => {
  const db = await mysql.createConnection(poolOption);
  await execScripts(db, path.join(import.meta.dirname, "./db.init.sql"));

  test.test("Indexing webpage", async (t) => {
    try {
      for (let w of webpages) {
        await dbInterface.saveWebpage(db, w);
      }
    } catch (e: any) {
      console.error(e);
      t.assert.fail(e.message);
    }
  });

  test.test("Check if websites have already been indexed", async (t) => {
    try {
      const l = await dbInterface.checkIndexedWebpage(db, [
        "https://youtube.com",
        "https://exudos.ai",
        "https://fedex.com",
      ]);
      t.assert.equal(3, l.length, "Not equal");
    } catch (e: any) {
      console.error(e);
      t.assert.fail(e);
    }
  });
});

test.suite("Frontier Queue", { only: true }, async () => {
  const db = await mysql.createConnection(poolOption);
  await execScripts(db, path.join(import.meta.dirname, "./db.init.sql"));

  const urls: URLs = {
    Root: "domain.com",
    Nodes: [
      "https://domain.com/",
      "https://domain.com/app",
      "https://domain.com/app/id",
      "https://domain.com/blogs",
      "https://domain.com/api/v2/person",
    ],
  };

  test.test("enqueuing nodes", async (t) => {
    try {
      await dbInterface.enqueueUrls(db, urls);
    } catch (e: any) {
      t.assert.fail(e);
    } finally {
      await db.end();
    }
  });
});
