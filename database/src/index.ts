import sqlite3 from "sqlite3";
import path from "path";
import amqp, { Connection } from "amqplib";
import channelOperations from "./rabbitmq/channel_operations";
import { readFile } from "fs";

const db = init_database();
const cumulativeAckCount = 1000;
exec_scripts(db, path.join(__dirname, "./db_utils/websites.init.sql"));

(async () => {
  console.log("Starting database server");
  try {
    const connection = await establishConnection(7);
    const databaseChannel = await connection.createChannel();
    console.log("Channel Created");
    databaseChannel.prefetch(cumulativeAckCount, false);
    await channelOperations.channelHandler(db, databaseChannel);
  } catch (err) {
    const error = err as Error;
    console.error(error.message);
  }
})();

async function establishConnection(retries: number): Promise<Connection> {
  if (retries > 0) {
    retries--;
    try {
      const connection = await amqp.connect("amqp://localhost:5672");
      console.log(
        `Successfully connected to rabbitmq after ${retries} retries`,
      );
      return connection;
    } catch (err) {
      console.error("Retrying Database service connection");
      await new Promise((resolve) => {
        const timeoutID = setTimeout(() => {
          resolve("Done blocking");
          clearTimeout(timeoutID);
        }, 2000);
      });
      return await establishConnection(retries);
    }
  }
  throw new Error("Shutting down database server after several retries");
}

function init_database(): sqlite3.Database {
  const dbFile = "../website_collection.db";
  const sqlite = sqlite3.verbose();
  const db = new sqlite.Database(
    path.join(__dirname, dbFile),
    sqlite.OPEN_READWRITE,
    (err) => {
      if (err) {
        console.error("Unable to connect to website_collection.db");
        process.exit(1);
      }
      console.log("Connected to sqlite3 database.");
    },
  );
  return db;
}

async function exec_scripts(db: sqlite3.Database, scriptPath: string) {
  console.log("Execute sqlite script");
  readFile(scriptPath, "utf-8", (err, data) => {
    const stmts = data
      .split(";")
      .map((stmt) => stmt.trim())
      .filter((stmt) => stmt);

    stmts.forEach((statement) => {
      db.run(statement, [], (err) => {
        console.log("Executed statement: %s ", statement);
      });
    });
  });
}
