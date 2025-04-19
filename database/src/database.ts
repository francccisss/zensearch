import type {
  IndexedWebpage,
  URLs,
  Webpage,
  FrontierQueue,
} from "./utils/types.ts";
import { randomUUID } from "crypto";
import Database from "better-sqlite3";

// TODO ADD UUID INSTEAD OF DATABASE INCREMENTING IDS
async function saveWebpage(db: Database.Database, data: IndexedWebpage) {
  if (db == null) {
    throw new Error("ERROR: Database is not connected.");
  }

  const insertIndexedSitesStmt = db.prepare(
    "INSERT INTO indexed_sites (id, primary_url, last_indexed) VALUES (?,?,?);",
  );

  const indexedSiteID = randomUUID();
  insertIndexedSitesStmt.run(
    indexedSiteID,
    new URL(data.URLSeed).hostname,
    Date.now(),
  );

  const insertWebpageStmt = db.prepare(
    "INSERT INTO webpages (url, title, contents, parent) VALUES (?, ?, ?, ?);",
  );
  insertWebpageStmt.run(
    data.Webpage.Header.Url,
    data.Webpage.Header.Title,
    data.Webpage.Contents,
    indexedSiteID,
  );

  console.log("NOTIF: DONE INDEXING");
}

async function queryWebpages(db: Database.Database): Promise<Array<Webpage>> {
  // query function returns once the promise has either been resolved
  // or rejected by the sqlite query call.

  const stmt = db.prepare("SELECT Url, Contents, Title FROM webpages");
  const pages = stmt.all();
  return pages as Array<Webpage>;
}

// TODO clean this wrapper thingy majig next time. :D
function checkAlreadyIndexedWebpage(
  db: Database.Database,
  crawlList: Array<string>,
): Array<string> {
  const stmt = db.prepare(
    "SELECT primary_url as url FROM indexed_sites WHERE url = ?",
  );

  if (crawlList.length === 0) return [];
  let tmp: Array<string> = crawlList.filter((item) => {
    const hostname = new URL(item).hostname;
    const webpage = stmt.get(hostname);
    return webpage != undefined;
  });

  return tmp;
}

function enqueueUrls(db: Database.Database, Urls: URLs) {
  // check if domain exists
  console.log("DATABASE NAME: %s", db.name);
  const stmt = db.prepare("SELECT * FROM queues WHERE domain = ?");
  let domain = stmt.get(Urls.Domain) as FrontierQueue | undefined;
  if (domain === undefined) {
    console.log(
      "Domain does not exist, creating a new domain from `%s`",
      Urls.Domain,
    );
    db.prepare("INSERT INTO queues (id, domain) VALUES (?, ?)").run(
      randomUUID(),
      Urls.Domain,
    );
    domain = stmt.get(Urls.Domain) as FrontierQueue;
    console.log("Domain created");
  }
  const nodeInsert = db.prepare(
    "INSERT INTO nodes (id, url, queue_id) VALUES (?, ?, ?)",
  );
  console.log("Inserting new nodes to queue");
  Urls.Nodes.forEach((node) => {
    nodeInsert.run(randomUUID(), node, domain.id);
  });
  console.log("Nodes Enqueued");
}

async function clearURLs(db: Database.Database, q: FrontierQueue) {
  console.log(q);
  console.log("Storing URLS in FrontierQueue");
}

// What is dequeued is considered a Visited Node, so
async function dequeueURL(
  db: Database.Database,
  src: string,
): Promise<{ length: number; url: string }> {
  return { length: 0, url: "fzaid.vercel.app/home" };
}
export default {
  saveWebpage,
  checkAlreadyIndexedWebpage,
  queryWebpages,
  enqueueUrls,
  clearURLs,
  dequeueURL,
};
