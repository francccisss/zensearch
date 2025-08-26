import path from "path";
import Database from "better-sqlite3";
import rabbitmq from "./rabbitmq/index.js";
import { readFile } from "fs";
import { exit } from "node:process";
import { readFileSync } from "node:fs";

const wc = path.join(import.meta.dirname, "../../website_collection.db");
const fq = path.join(import.meta.dirname, "../../frontier_queue.db");
const websitesDB = initDatabase(wc);
const frontierQueueDB = initDatabase(fq);
const cumulativeAckCount = 1000;

const tables = [
  "known_sites",
  "indexed_sites",
  "webpages",
  "visited_nodes", // Dont move this before node
  "nodes",
  "queues",
];
await (async (): Promise<void> => {
  try {
    await execScripts(
      websitesDB,
      path.join(import.meta.dirname, "./db_utils/websites.init.sql"),
    );
    await execScripts(
      frontierQueueDB,
      path.join(import.meta.dirname, "./db_utils/frontier_queue.sql"),
    );

    console.log("Notif: tables created");
    console.log("Starting database server");
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
  console.log(`Executing sql script for ${scriptPath}`);
  if (db === null) {
    console.error("ERROR: database does not exist for %s", scriptPath);
    exit(1);
  }

  readFile(scriptPath, "utf-8", (err, data) => {
    if (err !== null) {
      console.log(err);
      throw new Error(err.message);
    }
    const stmts = data.split(";").map((stmt) => stmt.trim());
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
