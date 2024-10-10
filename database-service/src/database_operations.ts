import { Database } from "sqlite3";
import { data_t, webpage_t } from "./utils/types";

async function index_webpages(db: Database, data: data_t) {
  if (db == null) {
    throw new Error("ERROR: Database is not connected.");
  }
  db.serialize(() => {
    db.run(
      "INSERT OR IGNORE INTO known_sites (url, last_added) VALUES ($url, $last_added);",
      {
        $url: data.Url,
        $last_added: Date.now(),
      },
    );
    const insert_indexed_sites_stmt = db.prepare(
      "INSERT OR IGNORE INTO indexed_sites (primary_url, last_indexed) VALUES ($primary_url, $last_indexed);",
    );
    const insert_webpages_stmt = db.prepare(
      "INSERT INTO webpages (url, title, contents, parent) VALUES ($webpage_url, $title, $contents, $parent);",
    );
    insert_indexed_sites_stmt.run(
      {
        $primary_url: data.Url,
        $last_indexed: Date.now(),
      },
      function (err) {
        if (err) {
          console.error("Unable to add last indexed site:", err.message);
          return;
        }
        const parentId = this.lastID;
        data.Webpages.forEach((el) => {
          if (el === undefined) return;
          const {
            Header: { Title, Url },
            Contents,
          } = el;

          insert_webpages_stmt.run(
            {
              $webpage_url: Url,
              $title: Title,
              $contents: Contents,
              $parent: parentId,
            },
            (err) => {
              if (err) {
                console.error("Error inserting webpage:", err.message);
              }
            },
          );
        });
        insert_webpages_stmt.finalize();
      },
    );
    insert_indexed_sites_stmt.finalize();
  });
  console.log("NOTIF: DONE INDEXING");
}

async function query_webpages(db: Database): Promise<Array<webpage_t>> {
  // query function returns once the promise has either been resolved
  // or rejected by the sqlite query call.

  return await new Promise(function (resolved, reject) {
    const sql_query = "SELECT Url, Contents, Title FROM webpages";
    db.all<webpage_t>(sql_query, (err, row) => {
      try {
        if (err) {
          throw new Error(
            "ERROR: Something went wrong while querying webpages for search query.",
          );
        }
        if (row.length === 0) {
          console.log("There are 0 webpages.\n");
          console.log("Please Crawl the web. \n");
        }
        resolved(row);
      } catch (err) {
        const error = err as Error;
        console.log(
          "ERROR: Error while querying webpages in database service.",
        );
        console.error(error.message);
        reject(error.message);
      }
    });
  });
}

// TODO clean this wrapper thingy majig next time. :D
async function check_existing_tasks(
  db: Database,
  crawl_list: Array<string>,
): Promise<Array<string>> {
  let tmp: Array<string> = [];
  const query = await query_promise_wrapper(db);
  if (query == null)
    throw new Error("ERROR: Unable to query indexed websites.");
  crawl_list.forEach((item) => {
    const url = new URL(item).hostname;
    if (!query.has(url)) {
      tmp.push(item);
    }
  });
  return tmp;
}

async function query_promise_wrapper(
  db: Database,
): Promise<Map<string, string> | null> {
  const stmt = `SELECT primary_url FROM indexed_sites`;
  let indexed_map: Map<string, string> = new Map();
  return new Promise((resolve, reject) => {
    db.each(
      stmt,
      function (err, row: { primary_url: string }) {
        try {
          if (err) {
            throw new Error(err.message);
          }
          indexed_map.set(row.primary_url, "");
        } catch (err) {
          console.error("ERROR: Unable to query indexed websites.");
          console.error(err);
          reject(null);
        }
      },
      function (err) {
        if (err) {
          console.error(err);
          reject(null);
          return;
        }
        resolve(indexed_map);
      },
    );
  });
}

export default { index_webpages, check_existing_tasks, query_webpages };
