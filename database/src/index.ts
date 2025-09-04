import path from "path";
import rabbitmq from "./rabbitmq/index.js";
import mysql from "mysql2/promise";
import "dotenv/config";
import { exit } from "node:process";
import { readFileSync } from "node:fs";

const cumulativeAckCount = 1000;

const poolOption: mysql.PoolOptions = {
  user: process.env.DB_USER,
  password: process.env.DB_PASS,
  database: process.env.DB_NAME,
  host: process.env.DB_HOST,
  multipleStatements: false,
};

export async function init(): Promise<void> {
  try {
    const db = mysql.createPool(poolOption);
    await execScripts(
      db,
      path.join(import.meta.dirname, "./db_utils/db.init.sql"),
    );

    console.log("Notif: tables created");
    console.log("Starting database server");
    const connection = await rabbitmq.establishConnection(7);
    const databaseChannel = await connection.createChannel();
    const frontierChannel = await connection.createChannel();
    console.log("Channel Created");
    databaseChannel.prefetch(cumulativeAckCount, false);
    rabbitmq.webpageHandler(db, databaseChannel);
    rabbitmq.frontierQueueHandler(db, frontierChannel);
  } catch (err) {
    const error = err as Error;
    console.error(error);
    exit(1);
  }
}

export async function execScripts(
  db: mysql.Pool | null,
  scriptPath: string,
): Promise<void> {
  console.log(`Executing sql script for ${scriptPath}`);
  if (db === null) {
    console.error("ERROR: database does not exist for %s", scriptPath);
    exit(1);
  }

  const f = readFileSync(scriptPath, "utf-8");
  // TODO: Make sure each table that references a table needs to be triggered right after the
  // referenced table has already been created
  try {
    f.split(";")
      .filter((t) => t.trim())
      .forEach(async (t) => {
        try {
          await db.execute(t);
        } catch (e: any) {
          if (e.message.includes("exists")) {
            console.log("skipping duplicate");
            return;
          }
          console.error(e);
          throw new Error(e);
        }
      });
  } catch (e: any) {
    throw new Error(e);
  }
}
