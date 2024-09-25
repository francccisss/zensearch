import amqp from "amqplib";
import database_operations from "../database_operations";
import { Database } from "sqlite3";
import data_serializer from "../utils/32bit_serializer";
import { stringify } from "querystring";
import { writeFileSync } from "fs";

async function channel_handler(db: Database, ...args: Array<amqp.Channel>) {
  const push_queue = "database_push_queue";
  const query_queue = "database_query_queue";
  const db_response_queue = "database_response_queue";
  const [push_channel, query_channel] = args;

  await push_channel.assertQueue(push_queue, {
    exclusive: false,
    durable: false,
  });
  await query_channel.assertQueue(query_queue, {
    exclusive: false,
    durable: false,
  });

  // TODO Document code please :)

  push_channel.consume(push_queue, async (data) => {
    if (data === null) throw new Error("No data was pushed.");
    const decoder = new TextDecoder();
    const decoded_data = decoder.decode(data.content as ArrayBuffer);
    try {
      const deserialize_data = JSON.parse(decoded_data);
      console.log(deserialize_data);
      database_operations.index_webpages(db, deserialize_data);
    } catch (err) {
      const error = err as Error;
      console.error("ERROR: Decoder was unable to deserialized indexed data.");
      console.error(error.message);
      console.error(error.stack);
    } finally {
      push_channel.ack(data);
    }
  });
  query_channel.consume(query_queue, async (data) => {
    if (data === null) throw new Error("No data was pushed.");
    try {
      const data_query: {
        Contents: string;
        Title: string;
        Url: string;
      }[] = await database_operations.query_webpages(db);
      console.log(data.content);
      await query_channel.sendToQueue(
        db_response_queue,
        Buffer.from(JSON.stringify(data_query)),
        {
          correlationId: data.properties.correlationId,
        },
      );
      query_channel.ack(data);
    } catch (err) {
      const error = err as Error;
      console.error(error.message);
      console.error(error.stack);
      query_channel.nack(data, false, false);
    }
  });
}

export default { channel_handler };
