import amqp from "amqplib";
import database_operations from "../database_operations";
import { Database } from "sqlite3";
import data_serializer from "../utils/32bit_serializer";

/*
  channel handler can take in multiple channels from a single tcp conneciton
  to the rabbitmq message broker, these channels are multiplexed to handle
  messages coming from different context eg: database and search
*/

async function channel_handler(db: Database, ...args: Array<amqp.Channel>) {
  const push_queue = "database_push_queue";

  // routing key used by search engine service to query database for webpages.
  const db_query_queue = "database_query_queue";
  // routing key to reply back to the search engine service's callback queue.
  const db_response_queue = "database_response_queue";

  // routing key used by express server to check existing webpages.
  const db_check_express = "db_check_express";
  const db_cbq_express = "db_cbq_express";

  const [push_channel, database_channel] = args;

  await push_channel.assertQueue(push_queue, {
    exclusive: false,
    durable: false,
  });

  /*
   Channel to consume messages sent by the search engine service.
   The fuck do you mean push_channel? this channel is supposed to consume
   the data from the web crawler, it receives an array of Type Webpages.

   TODO Document code please :)
   TODO Change names to make it more comprehensible please :D
  */

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

  // DATABASE MESSAGE HANDLERS
  await database_channel.assertQueue(db_query_queue, {
    exclusive: false,
    durable: false,
  });

  await database_channel.assertQueue(db_check_express, {
    exclusive: false,
    durable: false,
  });

  /*
    The `db_check_express`consumes the messages from the express server
    for new crawl tasks, its responsibility is to check send
    the array of crawl tasks from the client to the database service
    and query the database to see if the list of crawl tasks, exists
    on the database already by checking the indexed_sites TABLE,
    if it does it returns back the array filtering out the ones that
    have already been crawled and indexed into the database, returning
    the remaining items as uncrawled to be processed by the crawler service.
   */

  database_channel.consume(db_check_express, async (data) => {
    if (data == null) throw new Error("No data was pushed.");
    try {
      console.log(
        "NOTIF: DB service received crawl list to check existing indexed websites.",
      );
      const crawl_list: { Docs: Array<string> } = JSON.parse(
        data.content.toString(),
      );
      const unindexed_websites = await database_operations.check_existing_tasks(
        db,
        crawl_list.Docs,
      );

      console.log(unindexed_websites);
      const encoder = new TextEncoder();
      const encoded_docs = encoder.encode(
        JSON.stringify({ Docs: unindexed_websites }),
      );

      const is_sent = database_channel.sendToQueue(
        db_cbq_express,
        Buffer.from(encoded_docs),
      );
      if (!is_sent) {
        throw new Error("ERROR: Unable to send back message.");
      }
      database_channel.ack(data);
    } catch (err) {
      console.error(err);
    }
  });

  /*
    Consumes messages sent by the search engine services to query the database for
    the webpages to be ranked and sent to the client.

    a callback queue is implemented once we consume and query webpages so that we can send
    it back right after all those process are done.
  */
  database_channel.consume(db_query_queue, async (data) => {
    if (data === null) throw new Error("No data was pushed.");
    try {
      const data_query: {
        Contents: string;
        Title: string;
        Url: string;
      }[] = await database_operations.query_webpages(db);
      console.log(data.content);
      await database_channel.sendToQueue(
        db_response_queue, // respond back to this queue from search engine
        Buffer.from(JSON.stringify(data_query)),
        {
          correlationId: data.properties.correlationId,
        },
      );
      database_channel.ack(data);
    } catch (err) {
      const error = err as Error;
      console.error(error.message);
      console.error(error.stack);
      database_channel.nack(data, false, false);
    }
  });
}

export default { channel_handler };
