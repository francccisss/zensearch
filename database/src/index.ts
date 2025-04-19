import path from "path";
import Database from "better-sqlite3";
import rabbitmq from "./rabbitmq/index.ts";
import { readFile } from "fs";
import { exit } from "node:process";

const wc = path.join(import.meta.dirname, "../website_collection.db");
const fq = path.join(import.meta.dirname, "../frontier_queue.db");
const websitesDB = initDatabase(wc);
const frontierQueueDB = initDatabase(fq);
const cumulativeAckCount = 1000;
execScripts(
  websitesDB,
  path.join(import.meta.dirname, "./db_utils/websites.init.sql"),
);
execScripts(
  frontierQueueDB,
  path.join(import.meta.dirname, "./db_utils/frontier_queue.sql"),
);

const tables = [
  "known_sites",
  "indexed_sites",
  "webpages",
  "visited_node", // Dont move this before node
  "node",
  "queue",
];
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

function initDatabase(src: string): Database.Database {
  const db = new Database(src);
  return db;
}

async function execScripts(db: Database.Database | null, scriptPath: string) {
  console.log("Execute sqlite script");
  console.log(scriptPath);
  if (db === null) {
    console.error("ERROR: database does not exist for %s", scriptPath);
    exit(1);
  }
  readFile(scriptPath, "utf-8", (_, data) => {
    const stmts = data
      .split(";")
      .map((stmt) => stmt.trim())
      .filter((stmt) => stmt);
    stmts.forEach((stmt) => {
      const firstLine = stmt.split("\n")[0];
      for (let i = 0; i < tables.length; i++) {
        const tableName = tables[i];
        if (firstLine.includes(tableName)) {
          console.log(tableName);
          const checkTableStmt = db.prepare(
            `SELECT name FROM sqlite_master WHERE type='table' AND name=? ;`,
          );
          const result = checkTableStmt.get([tableName]);
          if (result == undefined) {
            console.log("Notif: Creating table for %s", tableName);
            db.exec(stmt);
            break;
          }
          console.log("Notif:  Table already exists for %s", tableName);
        }
      }
    });
  });
}
