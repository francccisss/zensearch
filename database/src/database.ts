import type {
  IndexedWebpage,
  URLs,
  Webpage,
  FrontierQueue,
  Node,
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
  // check if the current node has already been visited
  const nodeInsert = db.prepare(
    "INSERT INTO nodes (url, queue_id) VALUES (?, ?)",
  );
  console.log("Inserting new nodes to queue");
  Urls.Nodes.forEach((node) => {
    nodeInsert.run(node, domain.id);
  });
  console.log("Nodes Enqueued");
}

function clearURLs(db: Database.Database, q: FrontierQueue) {
  console.log(q);
  console.log("Storing URLS in FrontierQueue");
  console.log(db);
}

// What is dequeued is considered a Visited Node
// Qdomain is used to identify the queue that is in-used by the crawler
function dequeueURL(
  db: Database.Database,
  Qdomain: string,
): { length: number; url: string; message: string } {
  try {
    const stmt = db.prepare("SELECT * FROM queues WHERE domain = ?");
    let domain = stmt.get(Qdomain) as FrontierQueue | undefined;
    if (domain === undefined) {
      console.log("Domain does not exist = `%s`", Qdomain);
      throw new Error(`Domain does not exist = '${Qdomain}'`);
    }

    // used for continuing if the crawler crashes for some reason
    // if the node was not sent to the crawler
    // if the database fails
    const inProgressNode = db
      .prepare("SELECT * from nodes WHERE status = 'in_progress'")
      .get() as Node;

    if (inProgressNode == undefined) {
      const nextNode = db
        .prepare(
          "SELECT * FROM nodes WHERE status = 'pending' ORDER BY id LIMIT 1",
        )
        .get() as Node;

      if (nextNode === undefined) {
        db.prepare("DELETE FROM queues WHERE id = ?").run(domain.id);
        return {
          length: 0,
          url: "",
          message: `Frontier queue for ${domain.domain} is empty.`,
        };
      }

      db.prepare("UPDATE nodes SET status = 'in_progress' WHERE id = ?").run(
        nextNode.id,
      );
      console.log("Update node=%s to 'in_progess'", nextNode.id);
      const nodes = db
        .prepare("SELECT * from nodes WHERE status = 'pending'")
        .all() as Node[];

      return { length: nodes.length, url: nextNode.url, message: "" };
    }

    const nodes = db
      .prepare("SELECT * from nodes WHERE status = 'pending'")
      .all() as Node[];

    if (nodes.length == 0) {
      console.log("Queue is empty clean up queue.");
      db.prepare("DELETE FROM queues WHERE id = ?").run(domain.id);
      return {
        length: 0,
        url: "",
        message: `Frontier queue for ${domain.domain} is empty.`,
      };
    }

    console.log("Update node=%s to 'in_progess'", inProgressNode.id);

    return {
      length: nodes.length,
      url: inProgressNode.url,
      message: "",
    };
  } catch (e) {
    const err = e as Error;
    return { length: 0, url: "", message: err.message };
  }
}

function setNodeToVisited(db: Database.Database, node: Node) {
  try {
    db.prepare(
      "INSERT INTO visited_nodes (node_url, queue_id) VALUES (?, ?)",
    ).run(node.url, node.queue_id);
  } catch (err) {
    console.error("Error: Unable to set node as visited");
    console.error(err);
  }
}

function checkNodeVisited(db: Database.Database, url: string): boolean {
  return db
    .prepare(
      "SELECT * FROM visited_nodes vn JOIN nodes n ON ? = n.url JOIN queues ON vn.queue_id = queues.id",
    )
    .get(url) == undefined
    ? false
    : true;
}

export default {
  checkNodeVisited,
  saveWebpage,
  checkAlreadyIndexedWebpage,
  queryWebpages,
  enqueueUrls,
  clearURLs,
  dequeueURL,
  setNodeToVisited,
};
