import amqp from "amqplib";
import database_operations from "../database_operations";
import { Database } from "sqlite3";
import data_serializer from "../utils/32bit_serializer";
import { data_t } from "../utils/types";

/*
  channel handler can take in multiple channels from a single tcp conneciton
  to the rabbitmq message broker, these channels are multiplexed to handle
  messages coming from different context eg: database and search
*/

async function channel_handler(db: Database, ...args: Array<amqp.Channel>) {
  // CRAWLER ROUTING KEYS
  const db_indexing_crawler = "db_indexing_crawler";

  // SEARCH ENGINE ROUTING KEYS
  // routing key used by search engine service to query database for webpages.
  const db_query_sengine = "db_query_sengine";
  // routing key to reply back to the search engine service's callback queue.
  const db_cbq_sengine = "db_cbq_sengine";

  // EXPRESS SERVER ROUTING KEYS
  // routing key used by express server to check existing webpages.
  const db_check_express = "db_check_express";
  const db_cbq_express = "db_cbq_express";
  const db_cbq_poll_express = "crawl_poll_queue";

  const [database_channel] = args;

  /*
   TODO Document code please :)
   TODO Change names to make it more comprehensible please :D

    Consumer waits for the crawler service to push new webpages into the `db_indexing_crawler`
    message queue, the `index_webpages` handler saves these crawled webpages into
    the database.
  */
  await database_channel.assertQueue(db_indexing_crawler, {
    exclusive: false,
    durable: false,
  });

  // Aside from database service to handle errors for indexing,
  // the crawler also needs to send the same mssage format straight
  // to the express server to notify that the crawl for the current url
  // threw an error.

  database_channel.consume(db_indexing_crawler, async (data) => {
    if (data === null) throw new Error("No data was pushed.");
    const decoder = new TextDecoder();
    const decoded_data = decoder.decode(data.content as ArrayBuffer);
    const deserialize_data: data_t = JSON.parse(decoded_data);
    try {
      // if ever an error occurs while saving new indexed webpages.
      // send message to the client with a message of "Error" with the.
      // current url.
      database_channel.ack(data);
      // sqlite is asynchronous, after calling this function
      // it will call sqlite query function asynchronously, so this one should return
      // immediately
      await database_operations.index_webpages(db, deserialize_data);
      database_channel.sendToQueue(
        data.properties.replyTo,
        Buffer.from(
          JSON.stringify({
            Message: "Success",
            Url: deserialize_data.Url,
            WebpageCount: deserialize_data.Webpages.length,
          }),
        ),
      );
    } catch (err) {
      const error = err as Error;
      console.error("ERROR: Decoder was unable to deserialized indexed data.");
      console.error(error.message);
      database_channel.nack(data, false, false);
      database_channel.sendToQueue(
        data.properties.replyTo,
        Buffer.from(
          JSON.stringify({
            message: "Error",
            url: deserialize_data.Url,
            webpage_count: 0,
          }),
        ),
      );
    }
  });

  // SEARCH ENGINE AND EXPRESS JS CONSUMERS
  await database_channel.assertQueue(db_query_sengine, {
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

      database_channel.ack(data);
      const is_sent = database_channel.sendToQueue(
        db_cbq_express,
        Buffer.from(encoded_docs),
      );
      if (!is_sent) {
        throw new Error("ERROR: Unable to send back message.");
      }
    } catch (err) {
      const error = err as Error;
      database_channel.nack(data, false, false);
      console.error(error.message);
      console.error(err);
    }
  });

  /*
    Consumes messages sent by the search engine services to query the database for
    the webpages to be ranked and sent to the client.

    a callback queue is implemented once we consume and query webpages so that we can send
    it back right after all those process are done.
  */
  database_channel.consume(db_query_sengine, async (data) => {
    if (data === null) throw new Error("No data was pushed.");
    try {
      const data_query: {
        Contents: string;
        Title: string;
        Url: string;
      }[] = await database_operations.query_webpages(db);
      console.log(data.content);
      database_channel.ack(data);
      await database_channel.sendToQueue(
        db_cbq_sengine, // respond back to this queue from search engine
        Buffer.from(JSON.stringify(data_query)),
        {
          correlationId: data.properties.correlationId,
        },
      );
    } catch (err) {
      const error = err as Error;
      console.error(error.message);
      console.error(error.stack);
      database_channel.nack(data, false, false);
    }
  });
}

export default { channel_handler };
