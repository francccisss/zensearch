import { IndexedWebpage, URLs, Webpage, FrontierQueue } from "./utils/types";
import { randomUUID } from "crypto";
import Database from "better-sqlite3";

// TODO ADD UUID INSTEAD OF DATABASE INCREMENTING IDS
async function indexWebpages(db: Database.Database, data: IndexedWebpage) {
  if (db == null) {
    throw new Error("ERROR: Database is not connected.");
  }
  const insertKnownStmt = db.prepare(
    "INSERT INTO known_sites (id, url, last_added) VALUES (@id, @url, @last_added);",
  );

  insertKnownStmt.run({
    id: randomUUID(),
    url: data.URLSeed,
    last_added: Date.now(),
  });

  const insertIndexedSitesStmt = db.prepare(
    "INSERT INTO indexed_sites (id, primary_url, last_indexed) VALUES (@id ,@primary_url, @last_indexed);",
  );

  const indexedSiteID = randomUUID();
  insertIndexedSitesStmt.run({
    id: indexedSiteID,
    primary_url: data.URLSeed,
    last_indexed: Date.now(),
  });

  const insertWebpageStmt = db.prepare(
    "INSERT INTO webpages (url, title, contents, parent) VALUES (@webpage_url, @title, @contents, @parent);",
  );
  insertWebpageStmt.run({
    webpage_url: data.Webpage.Header.Url,
    title: data.Webpage.Header.Title,
    contents: data.Webpage.Contents,
    parent: indexedSiteID,
  });

  console.log("NOTIF: DONE INDEXING");
}

async function queryWebpages(db: Database.Database): Promise<Array<Webpage>> {
  // query function returns once the promise has either been resolved
  // or rejected by the sqlite query call.

  return await new Promise(function (resolved, reject) {
    //  const sql_query = "SELECT Url, Contents, Title FROM webpages";
    //  db.all(sql_query, (err: Error, rows: Array<Webpage>) => {
    //    try {
    //      if (err) {
    //        throw new Error(
    //          "ERROR: Something went wrong while querying webpages for search query.",
    //        );
    //      }
    //      if (rows.length === 0) {
    //        console.log("There are 0 webpages.\n");
    //        console.log("Please Crawl the web. \n");
    //      }
    //      resolved(rows);
    //    } catch (err) {
    //      const error = err as Error;
    //      console.log(
    //        "ERROR: Error while querying webpages in database service.",
    //      );
    //      console.error(error.message);
    //      reject(error.message);
    //    }
    //  });
  });
}

// TODO clean this wrapper thingy majig next time. :D
async function checkExistingTasks(
  db: Database.Database,
  crawlList: Array<string>,
): Promise<Array<string>> {
  let tmp: Array<string> = [];
  //const query = await queryPromiseWrapper(db);
  //if (query == null)
  //  throw new Error("ERROR: Unable to query indexed websites.");
  //crawlList.forEach((item) => {
  //  const url = new URL(item).hostname;
  //  if (!query.has(url)) {
  //    tmp.push(item);
  //  }
  //});
  return tmp;
}

async function queryPromiseWrapper(
  db: Database.Database,
): Promise<Map<string, string> | null> {
  const stmt = `SELECT primary_url FROM indexed_sites`;
  let indexed_map: Map<string, string> = new Map();
  return new Promise((resolve, reject) => {
    //db.each(
    //  stmt,
    //  function (err: Error, row: { primary_url: string }) {
    //    try {
    //      if (err) {
    //        throw new Error(err.message);
    //      }
    //      indexed_map.set(row.primary_url, "");
    //    } catch (err) {
    //      console.error("ERROR: Unable to query indexed websites.");
    //      console.error(err);
    //      reject(null);
    //    }
    //  },
    //  function (err) {
    //    if (err) {
    //      console.error(err);
    //      reject(null);
    //      return;
    //    }
    //    resolve(indexed_map);
    //  },
    //);
  });
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
  indexWebpages,
  checkExistingTasks,
  queryWebpages,
  storeURLs,
  clearURLs,
  dequeueURL,
};
