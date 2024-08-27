import sqlite3 from "sqlite3";
import path from "path";

function init_database() {
  const db_file = "./db_utils/websites.init.sql";
  console.log(db_file);
  const sqlite = sqlite3.verbose();
  const db = new sqlite.Database(
    path.join(__dirname, db_file),
    sqlite.OPEN_READWRITE,
    (err) => {
      if (err) {
        console.error("Unable to connect to website_collection.db");
        process.exit(1);
      }
      console.log("Thread connected to database");
    },
  );
  return db;
}
init_database();
