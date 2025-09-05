import type {
  IndexedWebpage,
  URLs,
  Webpage,
  FrontierQueue,
  Node,
  IndexedSite,
} from "./utils/types.js";
import mysql from "mysql2/promise";
import { randomUUID } from "crypto";

// URLSeed/primary_url and the url (webpages) from crawler should be different
// the URLSeed should be the entry point of the crawler, while the primary_url
// is the  that the contents correponds to.

async function saveWebpage(
  pool: mysql.Pool | mysql.Connection,
  data: IndexedWebpage,
): Promise<void> {
  if (pool == null) {
    throw new Error("ERROR: Database is not connected.");
  }

  const [indexedSiteQuery] = await pool.execute(
    "SELECT * FROM indexed_sites WHERE hostname = ?",
    [data.URLSeed],
  );
  const [indexedSite] = indexedSiteQuery as unknown as Array<
    IndexedSite | undefined
  >;

  const [queueQuery] = await pool.execute(
    "SELECT * FROM queues WHERE root = ?",
    [data.URLSeed],
  );
  const [queue] = queueQuery as unknown as Array<FrontierQueue | undefined>;

  // A queue is CREATED BEFORE saveWebpage is even called
  // since the crawler needs to DEQUEUE a url in the frontier QUEUE
  // before it navigates the webpage and will only save if there IS a QUEUE
  if (indexedSite !== undefined && queue == undefined) {
    console.log("Page fully indexed");
    return;
  }

  if (indexedSite !== undefined) {
    try {
      await pool.execute(
        "INSERT INTO webpages (id, url, title, contents, parent) VALUES (?,?, ?, ?, ?);",
        [
          randomUUID(),
          data.Webpage.Header.URL,
          data.Webpage.Header.Title,
          data.Webpage.Contents,
          indexedSite.id,
        ],
      );
      return;
    } catch (e: any) {
      if (e.message.toLowerCase().includes("duplicate")) {
        console.log(
          "WEBPAGE URL ALREADY EXISTS, SKIPPING=%s",
          data.Webpage.Header.URL,
        );
        return;
      }
      throw new Error(e);
    }
  }

  try {
    let indexedSiteID = randomUUID();
    await pool.execute(
      "INSERT INTO indexed_sites (id, hostname) VALUES (?,?);",
      [indexedSiteID, data.URLSeed],
    );
    await pool.execute(
      "INSERT INTO webpages (id, url, title, contents, parent) VALUES (?,?, ?, ?, ?);",
      [
        randomUUID(),
        data.Webpage.Header.URL,
        data.Webpage.Header.Title,
        data.Webpage.Contents,
        indexedSiteID,
      ],
    );
    console.log("NOTIF: DONE INDEXING");
  } catch (e: any) {
    throw new Error(e);
  }
}

async function queryWebpages(
  pool: mysql.Pool | mysql.Connection,
): Promise<Array<Webpage>> {
  // query function returns once the promise has either been resolved
  // or rejected by the sqlite query call.
  const [webpagesQuery] = await pool.execute(
    "SELECT url, contents, title FROM webpages",
  );
  const pages = webpagesQuery;
  return pages as Array<Webpage>;
}

// TODO clean this wrapper thingy majig next time. :D
async function checkIndexedWebpage(
  pool: mysql.Pool | mysql.Connection,
  crawlList: Array<string>,
): Promise<Array<string>> {
  if (crawlList.length === 0) return [];

  let tmp: Array<string> = [];

  for (let cl of crawlList) {
    console.log(cl);
    try {
      const [indexedSiteQuery] = await pool.execute(
        "SELECT hostname FROM indexed_sites WHERE hostname = ?",
        [new URL(cl).hostname],
      );
      const webpages = indexedSiteQuery as unknown as Array<IndexedSite>;
      if (webpages.length == 0) {
        tmp.push(cl);
      }
    } catch (e: any) {
      // making sure that we dont submit all the array and insaatead return an error
      throw new Error(e);
    }
  }
  console.log("WTASFDSASD");
  return tmp;
}

async function enqueueUrls(pool: mysql.Pool | mysql.Connection, Urls: URLs) {
  const [indexedSiteQuery] = await pool.execute(
    "SELECT * FROM indexed_sites WHERE hostname = ?",
    [Urls.Root],
  );
  const [indexedSite] = indexedSiteQuery as unknown as Array<
    IndexedSite | undefined
  >;

  const [queueQuery] = await pool.execute(
    "SELECT * FROM queues WHERE root = ?",
    [Urls.Root],
  );
  let queue = (queueQuery as unknown as Array<FrontierQueue>)[0] as
    | FrontierQueue
    | undefined;

  // A queue is CREATED BEFORE saveWebpage is even called
  // since the crawler needs to DEQUEUE a url in the frontier QUEUE
  // before it navigates the webpage and will only save if there IS a QUEUE
  if (indexedSite !== undefined && queue == undefined) {
    console.log("Page fully indexed");
    return;
  }

  if (queue === undefined) {
    console.log(
      "Queue does not exist, creating a new queue from `%s`",
      Urls.Root,
    );
    const newID = randomUUID();
    try {
      await pool.execute("INSERT INTO queues (id, root) VALUES (?, ?)", [
        newID,
        Urls.Root,
      ]);
      const [queueQuery] = await pool.execute(
        "SELECT * FROM queues WHERE id = ?",
        [newID],
      );
      queue = (queueQuery as unknown as Array<FrontierQueue>)[0] as
        | FrontierQueue
        | undefined;
      if (queue == undefined) {
        throw new Error(`Queue for ${Urls.Root} was not created!`);
      }
      console.log("Queue created");
    } catch (e: any) {
      throw new Error(e);
    }
  }
  // check if the current node has already been visited
  console.log("Inserting new nodes to queue");
  for (let node of Urls.Nodes) {
    // need to skip if it already exists
    if (!(await checkNodeVisited(pool, node))) {
      console.log("CHECKING IF EXIST/VISITED: %s", node);
      // Try catch prevent from crashing due to duplication of node.url
      try {
        await pool.execute("INSERT INTO nodes (url, queue_id) VALUES (?, ?)", [
          node,
          queue!.id,
        ]);
      } catch (e: any) {
        // a just in case there might be duplications
        if (e.message.toLowerCase().includes("duplicate")) {
          console.log("NODE ALREADY EXISTS, SKIPPING=%s", node);
          continue;
        }
        throw new Error(e);
      }
    } else {
      console.log("NODE ALREADY VISITED: ", node);
    }
  }
}

async function checkNodeVisited(
  pool: mysql.Pool | mysql.Connection,
  url: string,
): Promise<boolean> {
  const [query] = await pool.execute(
    "SELECT * FROM visited_nodes vn JOIN queues ON vn.queue_id = queues.id WHERE vn.node_url = ?",
    [url],
  );
  const q = query as unknown as Array<any>;
  return q.length > 0 ? true : false;
}

async function checkNodeExists(
  pool: mysql.Pool | mysql.Connection,
  url: string,
): Promise<boolean> {
  const [query] = await pool.execute(
    "SELECT * FROM nodes WHERE nodes.url = ?",
    [url],
  );
  const nodeRow = query as unknown as Array<Node>;
  return nodeRow.length > 0 ? true : false;
}

async function dequeueURL(
  pool: mysql.Pool | mysql.Connection,
  root: string,
): Promise<{
  length: number;
  url: string;
  message: string;
  inProgressNode: Node | null;
}> {
  try {
    const [queueQuery] = await pool.execute(
      "SELECT * FROM queues WHERE root = ?",
      [root],
    );

    const [queue] = queueQuery as unknown as Array<FrontierQueue | undefined>;

    if (queue === undefined) {
      return {
        length: 0,
        url: "",
        message: `Queue does not exist = '${root}'`,
        inProgressNode: null,
      };
    }

    // USED FOR RESUMING
    // if the crawler crashes for some reason
    // if the node was not sent to the crawler
    // if the database fails

    const [inProgressNodeQuery] = await pool.execute(
      "SELECT * from nodes WHERE status = 'in_progress' AND queue_id = ?",
      [queue.id],
    );
    let [inProgressNode] = inProgressNodeQuery as unknown as Array<Node>;

    if (inProgressNode === undefined) {
      const [nextNodeQuery] = await pool.execute(
        "SELECT * FROM nodes WHERE status = 'pending' and queue_id = ? ORDER BY id LIMIT 1",
        [queue.id],
      );
      const [nextNode] = nextNodeQuery as unknown as Array<Node | undefined>;

      if (nextNode === undefined) {
        return {
          length: 0,
          url: "",
          message: `Frontier queue for ${queue.root} is empty.`,
          inProgressNode: null,
        };
      }

      await pool.execute(
        "UPDATE nodes SET status = 'in_progress' WHERE id = ? AND queue_id = ?",
        [nextNode.id, queue.id],
      );
      nextNode.status = "in_progress"; // just so i dont get confused when logging it
      inProgressNode = nextNode;
    }

    // WTF IS THIS SHIITTIAISFAS
    const [nodeCountQuery] = await pool.execute(
      "SELECT COUNT(*) from nodes WHERE queue_id = ?",
      [queue.id],
    );

    const [nodeCount] = nodeCountQuery as unknown as Array<{
      [key: string]: any;
    }>;

    console.log("Update node=%s to 'in_progess'", inProgressNode.id);

    return {
      length: Object.values(nodeCount as { [key: string]: any })[0] as number,
      url: inProgressNode.url,
      message: "",
      inProgressNode: inProgressNode,
    };
  } catch (e) {
    const err = e as Error;
    console.error(err);
    return { length: 0, url: "", message: err.message, inProgressNode: null };
  }
}

// TODO: Change this when a user has sent the dequeued url
// when a user sends an enqueue to the specific queue, then we know that the
// indexing was successful

async function setNodeToVisited(
  pool: mysql.Pool | mysql.Connection,
  node: Node,
) {
  try {
    await pool.query("START TRANSACTION");
    await pool.execute(
      "INSERT INTO visited_nodes (node_url, queue_id) VALUES (?, ?)",
      [node.url, node.queue_id],
    );
    await pool.execute("DELETE FROM nodes WHERE nodes.id = ?", [node.id]);

    await pool.query("COMMIT");
  } catch (e: any) {
    await pool.query("ROLLBACK");
    console.error("Error: Unable to set node as visited");
    throw new Error(e);
  }
}

async function removeQueue(pool: mysql.Pool | mysql.Connection, root: string) {
  try {
    await pool.execute("DELETE FROM queues WHERE queues.root = ?", [root]);
  } catch (e: any) {
    console.error("ERROR: Unable to remove queue");
    throw new Error(e);
  }
}

async function getCurrentQueueLen(
  pool: mysql.Pool | mysql.Connection,
  root: string,
): Promise<number> {
  const [nodes] = await pool.execute(
    "SELECT * FROM (SELECT * FROM queues WHERE root = ?) AS cq, nodes WHERE cq.id = nodes.queue_id",
    [root],
  );
  return (nodes as unknown as Array<Node>).length;
}

export default {
  saveWebpage,
  checkNodeVisited,
  checkIndexedWebpage,
  queryWebpages,
  enqueueUrls,
  dequeueURL,
  setNodeToVisited,
  getCurrentQueueLen,
  removeQueue,
  checkNodeExists,
};
