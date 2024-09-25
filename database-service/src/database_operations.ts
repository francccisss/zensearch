import { Database } from "sqlite3";
import { data_t, webpage_t } from "./utils/types";

function index_webpages(db: Database, data: data_t) {
  if (db == null) {
    throw new Error("Database is not connected.");
  }
  console.log("INDEX PAGES");
  db.serialize(() => {
    // this.db.run("PRAGMA foreign_keys = ON;");
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
  console.log("DONE INDEXING");
}

async function query_webpages(db: Database): Promise<Array<webpage_t>> {
  return await new Promise(function (resolved, reject) {
    const sql_query = "SELECT Url, Contents, Title FROM webpages";
    db.all<webpage_t>(sql_query, (err, row) => {
      try {
        if (err) {
          throw new Error(
            "Something went wrong whil querying webpages for search query.",
          );
        }
        if (row.length === 0) {
          console.log("There are 0 webpages.\n");
          console.log("Please Crawl the web. \n");
        }
        resolved(row);
      } catch (err) {
        const error = err as Error;
        console.log("LOG: Error while querying webpages in database service.");
        console.error(error.message);
        reject(error.message);
      }
    });
  });
}

export default { index_webpages, query_webpages };
