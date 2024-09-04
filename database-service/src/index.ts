import sqlite3 from "sqlite3";
import path from "path";
import amqp from "amqplib";
import database_operations from "./database_operations";

const db = init_database();
(async () => {
  const connection = await amqp.connect("amqp://localhost");
  const push_channel = await connection.createChannel();
  const query_channel = await connection.createChannel();

  await handle_channels(push_channel, query_channel);
})();

async function handle_channels(...args: Array<amqp.Channel>) {
  const push_queue = "database_push_queue";
  const query_queue = "database_query_queue";
  const [push_channel, query_channel] = args;

  await push_channel.assertQueue(push_queue, {
    exclusive: false,
    durable: false,
  });
  await query_channel.assertQueue(query_queue, {
    exclusive: false,
    durable: false,
  });

  push_channel.consume(
    push_queue,
    async (data) => {
      if (data === null) throw new Error("No data was pushed.");
      console.log(data);
      const decoder = new TextDecoder();
      const decoded_data = decoder.decode(data.content as ArrayBuffer);
      const last_brace_index = decoded_data.lastIndexOf("}");
      const sliced_object = decoded_data.slice(0, last_brace_index) + "}"; // Buh
      try {
        const deserialize_data = JSON.parse(sliced_object);
        database_operations.index_webpages(db, deserialize_data);
      } catch (err) {
        const error = err as Error;
        console.log("LOG: Decoder was unable to deserialized indexed data.");
        console.error(error.message);
        console.error(error.stack);
      }
    },

    { noAck: false },
  );

  query_channel.consume(
    query_queue,
    async (data) => {
      try {
        if (data === null) throw new Error("No data was pushed.");
        console.log(data.properties.replyTo);
        console.log(data.content.toString("utf8"));
        await query_channel.sendToQueue(
          data.properties.replyTo,
          Buffer.from("Success Query"),
        );
      } catch (err) {
        const error = err as Error;
        console.error(error.message);
        console.error(error.stack);
      }
    },
    { noAck: false },
  );
}

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
