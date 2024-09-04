import { Database } from "sqlite3";
import { data_t } from "./utils/types";

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
        $url: new URL("/", data.header.url).hostname,
        $last_added: Date.now(),
      },
    );
    const insert_indexed_sites_stmt = db.prepare(
      "INSERT OR IGNORE INTO indexed_sites (primary_url, last_indexed) VALUES ($primary_url, $last_indexed);",
    );
    const insert_webpages_stmt = db.prepare(
      "INSERT INTO webpages (webpage_url, title, contents, parent) VALUES ($webpage_url, $title, $contents, $parent);",
    );
    insert_indexed_sites_stmt.run(
      {
        $primary_url: new URL("/", data.header.url).hostname,
        $last_indexed: Date.now(),
      },
      function (err) {
        if (err) {
          console.error("Unable to add last indexed site:", err.message);
          return;
        }
        const parentId = this.lastID;
        data.webpages.forEach((el) => {
          if (el === undefined) return;
          const {
            header: { title, webpage_url },
            contents,
          } = el;

          insert_webpages_stmt.run(
            {
              $webpage_url: webpage_url,
              $title: title,
              $contents: contents,
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

export default { index_webpages };
