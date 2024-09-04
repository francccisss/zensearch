import sqlite3 from "sqlite3";
import path from "path";
import amqp from "amqplib";
import database_operations from "./database_operations";
import channel_operations from "./rabbitmq/channel_operations";

const db = init_database();
(async () => {
  const connection = await amqp.connect("amqp://localhost");
  const push_channel = await connection.createChannel();
  const query_channel = await connection.createChannel();
  await channel_operations.channel_handler(db, push_channel, query_channel);
})();

function init_database(): sqlite3.Database {
  const db_file = "./website_collection.db";
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
