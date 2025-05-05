import type {
  IndexedWebpage,
  URLs,
  Webpage,
  FrontierQueue,
  Node,
  IndexedSite,
} from "./utils/types.js";
import { randomUUID } from "crypto";
import Database from "better-sqlite3";

// URLSeed/primary_url and the url (webpages) from crawler should be different
// the URLSeed should be the entry point of the crawler, while the primary_url
// is the  that the contents correponds to.

async function saveWebpage(db: Database.Database, data: IndexedWebpage) {
  if (db == null) {
    throw new Error("ERROR: Database is not connected.");
  }

  const indexedSite = db
    .prepare("SELECT * FROM indexed_sites WHERE primary_url = ?")
    .get(data.URLSeed) as IndexedSite | undefined;

  if (indexedSite !== undefined) {
    try {
      const insertWebpageStmt = db.prepare(
        "INSERT INTO webpages (url, title, contents, parent) VALUES (?, ?, ?, ?);",
      );
      insertWebpageStmt.run(
        data.Webpage.Header.URL,
        data.Webpage.Header.Title,
        data.Webpage.Contents,
        indexedSite.id,
      );
    } catch (err) {
      console.error(
        "WEBPAGE URL ALREADY EXISTS, SKIPPING=%s",
        data.Webpage.Header.URL,
      );
    }
    return;
  }

  let indexedSiteID = randomUUID();
  const insertIndexedSitesStmt = db.prepare(
    "INSERT INTO indexed_sites (id, primary_url, last_indexed) VALUES (?,?,?);",
  );

  insertIndexedSitesStmt.run(indexedSiteID, data.URLSeed, Date.now());
  const insertWebpageStmt = db.prepare(
    "INSERT INTO webpages (url, title, contents, parent) VALUES (?, ?, ?, ?);",
  );
  insertWebpageStmt.run(
    data.Webpage.Header.URL,
    data.Webpage.Header.Title,
    data.Webpage.Contents,
    indexedSiteID,
  );

  console.log("NOTIF: DONE INDEXING");
}

async function queryWebpages(db: Database.Database): Promise<Array<Webpage>> {
  // query function returns once the promise has either been resolved
  // or rejected by the sqlite query call.

  const stmt = db.prepare("SELECT url, contents, title FROM webpages");
  const pages = stmt.all();
  return pages as Array<Webpage>;
}

// TODO clean this wrapper thingy majig next time. :D
function checkAlreadyIndexedWebpage(
  db: Database.Database,
  crawlList: Array<string>,
): Array<string> {
  console.log("checking");
  const stmt = db.prepare(
    "SELECT primary_url as url FROM indexed_sites WHERE url = ?",
  );

  if (crawlList.length === 0) return [];
  let tmp: Array<string> = crawlList.filter((item) => {
    const hostname = new URL(item).hostname;
    const webpage = stmt.get(hostname);
    return webpage == undefined;
  });

  return tmp;
}

function enqueueUrls(db: Database.Database, Urls: URLs) {
  // check if domain exists
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
    // need to skip if it already exists
    if (!checkNodeVisited(db, node)) {
      console.log("CHECKING IF EXIST/VISITED: %s", node);
      // Try catch prevent from crashing due to duplication of node.url
      try {
        nodeInsert.run(node, domain.id);
      } catch (err) {
        console.log("NODE EXISTS: ", node);
      }
    } else {
      console.log("NODE VISITED: ", node);
    }
  });
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
): {
  length: number;
  url: string;
  message: string;
  inProgressNode: Node | null;
} {
  try {
    const stmt = db.prepare("SELECT * FROM queues WHERE domain = ?");
    let domain = stmt.get(Qdomain) as FrontierQueue | undefined;
    if (domain === undefined) {
      console.log("Domain does not exist = `%s`", Qdomain);
      return {
        length: 0,
        url: "",
        message: `Domain does not exist = '${Qdomain}'`,
        inProgressNode: null,
      };
    }

    // used for continuing if the crawler crashes for some reason
    // if the node was not sent to the crawler
    // if the database fails
    let inProgressNode = db
      .prepare("SELECT * from nodes WHERE status = 'in_progress'")
      .get() as Node;

    if (inProgressNode == undefined) {
      const nextNode = db
        .prepare(
          "SELECT * FROM nodes WHERE status = 'pending' ORDER BY id LIMIT 1",
        )
        .get() as Node;

      if (nextNode === undefined) {
        return {
          length: 0,
          url: "",
          message: `Frontier queue for ${domain.domain} is empty.`,
          inProgressNode: null,
        };
      }

      db.prepare("UPDATE nodes SET status = 'in_progress' WHERE id = ?").run(
        nextNode.id,
      );
      nextNode.status = "in_progress"; // just so i dont get confused when logging it
      inProgressNode = nextNode;
    }

    const nodeCount = db.prepare("SELECT COUNT(*) from nodes").get();

    // TODO FIX THIS, FOREIGN KEY CONSTRAINT, NODES RELYING ON REMOVED QUEUE
    // RETURNS ERROR, THE NODE WHERE ITS STATUS IS 'PENDING' STILL EXISTS
    // IN THE QUEUE, NEED TO MAKE SURE THE NODE IS REMOVED AFTER DEQUEUE IS CALLED
    // AND ONLY THEN WILL THE QUEUE BE REMOVED AS WELL

    console.log("Update node=%s to 'in_progess'", inProgressNode.id);

    return {
      length: Object.values(nodeCount as { [key: string]: any })[0] as number,
      url: inProgressNode.url,
      message: "",
      inProgressNode: inProgressNode,
    };
  } catch (e) {
    const err = e as Error;
    return { length: 0, url: "", message: err.message, inProgressNode: null };
  }
}

// only set once an ack has been received
function setNodeToVisited(db: Database.Database, node: Node) {
  try {
    db.transaction(() => {
      db.prepare("DELETE FROM nodes WHERE nodes.id = ?").run(node.id);
      db.prepare(
        "INSERT INTO visited_nodes (node_url, queue_id) VALUES (?, ?)",
      ).run(node.url, node.queue_id);
    })();
  } catch (err) {
    console.error("Error: Unable to set node as visited");
    console.error(err);
  }
}

function checkNodeExists(db: Database.Database, url: string): boolean {
  return db.prepare("SELECT * FROM nodes WHERE nodes.url = ?").get(url) ===
    undefined
    ? false
    : true;
}

function checkNodeVisited(db: Database.Database, url: string): boolean {
  return db
    .prepare(
      "SELECT * FROM visited_nodes vn JOIN queues ON vn.queue_id = queues.id WHERE vn.node_url = ?",
    )
    .get(url) == undefined
    ? false
    : true;
}

function removeQueue(db: Database.Database, domain: string) {
  try {
    db.prepare("DELETE FROM queues WHERE queues.domain = ?").run(domain);
  } catch (err) {
    console.error("ERROR: Unable to remove queue");
  }
}

function getCurrentQueueLen(db: Database.Database, domain: string): number {
  const nodes = db
    .prepare(
      "SELECT * FROM (SELECT * FROM queues WHERE domain = ?) AS cq, nodes WHERE cq.id = nodes.queue_id",
    )
    .all(domain) as Array<Node>;

  // if doesnt exist, just return 0, because the crawler will enqueue
  // a new url and set its corresponding queue to be used
  if (nodes == undefined) return 0;
  console.log(nodes);

  return nodes.length;
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
  getCurrentQueueLen,
  removeQueue,
  checkNodeExists,
};
