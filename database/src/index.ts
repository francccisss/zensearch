import path from "path";
import Database from "better-sqlite3";
import rabbitmq from "./rabbitmq";
import { readFile } from "fs";
import { exit } from "node:process";

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
    if (websitesDB == null || frontierQueueDB == null) {
      // TODO need to tell users which database initialization failed
      console.error("ERROR: unable to initialize databases");
      exit(1);
    }
    rabbitmq.webpageHandler(websitesDB, databaseChannel);
    rabbitmq.frontierQueueHandler(frontierQueueDB, frontierChannel);
  } catch (err) {
    const error = err as Error;
    console.error(error.message);
  }
})();

function init_database(src: string): Database.Database | null {
  try {
    const db = new Database(src);
    return db;
  } catch (err: any) {
    console.error("ERROR: Unable to initialize database %s", err.message);
    console.error(err);
    return null;
  }
}

async function exec_scripts(db: Database.Database | null, scriptPath: string) {
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
      db?.exec(stmt);
    });
  });
}
