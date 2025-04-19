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

const testDb: Map<string, Array<string>> = new Map();
async function storeURLs(db: Database.Database, Urls: URLs) {
  //db.serialize(() => {
  //  db.get(
  //    "SELECT Domain FROM Queue WHERE Domain = @Domain;",
  //    { @Domain: Urls.Domain },
  //    (err: Error, row: Array<FrontierQueue> | FrontierQueue) => {
  //      if (err != null) {
  //        console.error(err);
  //        throw new Error(err.message);
  //      }
  //      if (row == undefined) {
  //        db.run("INSERT INTO Queue (ID,Domain) VALUES (@ID, @Domain);", {
  //          @ID: randomUUID(),
  //          @Domain: Urls.Domain,
  //        });
  //      }
  //      Urls.Nodes.forEach((url) => {
  //        db.run(
  //          "INSERT INTO Node (ID, Url, QueueID) VALUES (@ID, @Url, @QueueID);",
  //          { @ID: randomUUID(), @Url: url, @QueueID: "21" },
  //        );
  //      });
  //    },
  //  );
  //});
  //
  //if (!testDb.has(Urls.Domain)) {
  //  console.log("Creating queue for %s", Urls.Domain);
  //  console.log("Stored URLS in FrontierQueue for %s", Urls.Domain);
  //  testDb.set(Urls.Domain, Urls.Nodes);
  //  return;
  //}
  //testDb.set(Urls.Domain, Urls.Nodes);
  //console.log("Stored URLS in FrontierQueue for %s", Urls.Domain);
}

async function clearURLs(db: Database.Database, q: FrontierQueue) {
  console.log(q);
  console.log("Storing URLS in FrontierQueue");
}

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
  storeURLs,
  clearURLs,
  dequeueURL,
};
