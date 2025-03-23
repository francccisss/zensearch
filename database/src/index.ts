import path from "path";
import sqlite3 from "sqlite3";
import rabbitmq from "./rabbitmq";
import { readFile } from "fs";

const wc = "../website_collection.db";
const fq = "../frontier_queue.db";
const websitesDB = init_database(wc);
const frontierQueueDB = init_database(fq);
const cumulativeAckCount = 1000;
exec_scripts(websitesDB, path.join(__dirname, "./db_utils/websites.init.sql"));
exec_scripts(
  frontierQueueDB,
  path.join(__dirname, "./db_utils/frontier_queue.sql"),
);

(async () => {
  console.log("Starting database server");
  try {
    const connection = await rabbitmq.establishConnection(7);
    const databaseChannel = await connection.createChannel();
    const frontierChannel = await connection.createChannel();
    console.log("Channel Created");
    databaseChannel.prefetch(cumulativeAckCount, false);
    rabbitmq.webpageHandler(websitesDB, databaseChannel);
    rabbitmq.frontierQueueHandler(frontierQueueDB, frontierChannel);
  } catch (err) {
    const error = err as Error;
    console.error(error.message);
  }
})();

function init_database(src: string): sqlite3.Database {
  const sqlite = sqlite3.verbose();
  const db = new sqlite.Database(
    path.join(__dirname, src),
    sqlite.OPEN_READWRITE,
    (err: Error | null) => {
      if (err != null) {
        console.error(err);
        console.error(err.message);
        console.error(
          "Unable to connect to website_collection.db make sure it exists first",
        );
        process.exit(1);
      }
      console.log("Connected to sqlite3 database.");
    },
  );
  return db;
}

async function exec_scripts(db: sqlite3.Database, scriptPath: string) {
  console.log("Execute sqlite script");
  readFile(scriptPath, "utf-8", (_, data) => {
    const stmts = data
      .split(";")
      .map((stmt) => stmt.trim())
      .filter((stmt) => stmt);

    stmts.forEach((statement) => {
      db.run(statement, [], (_) => {
        console.log("Executed statement: %s ", statement);
      });
    });
  });
}
