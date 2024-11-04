import sqlite3 from "sqlite3";
import path from "path";
import amqp from "amqplib";
import database_operations from "./database_operations";
import channel_operations from "./rabbitmq/channel_operations";
import { readFile } from "fs";

const db = init_database();
exec_scripts(db, path.join(__dirname, "./db_utils/websites.init.sql"));

(async () => {
  const connection = await amqp.connect("amqp://rabbitmq");
  console.log("Connected to rabbitmq");
  try {
    const database_channel = await connection.createChannel();
    console.log("Channel Created");
    await channel_operations.channel_handler(db, database_channel);
  } catch (err) {
    const error = err as Error;
    console.error(error.message);
    console.log("ERROR: Unable to create a channel for databas service.");
  }
})();

function init_database(): sqlite3.Database {
  const db_file = "./website_collection.db";
  const sqlite = sqlite3.verbose();
  const db = new sqlite.Database(
    path.join(__dirname, db_file),
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
  await readFile(scriptPath, "utf-8", (err, data) => {
    const stmts = data
      .split(";")
      .map((stmt) => stmt.trim())
      .filter((stmt) => stmt);

    stmts.forEach((statement) => {
      db.run(statement, [], (err) => {
        if (err) {
          console.error(err);
        }
        console.log("Executed statement: %s ", statement);
      });
    });
  });
}
