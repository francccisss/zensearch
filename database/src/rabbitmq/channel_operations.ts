import amqp from "amqplib";
import database_operations from "../database_operations";
import { Database } from "sqlite3";
import { Message, Webpage } from "../utils/types";
import segment_serializer from "../utils/segment_serializer";

/*
  channel handler can take in multiple channels from a single tcp conneciton
  to the rabbitmq message broker, these channels are multiplexed to handle
  messages coming from different context eg: database and search
*/

async function channel_handler(db: Database, database_channel: amqp.Channel) {
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
    const deserialize_data: Message = JSON.parse(decoded_data);
    try {
      database_channel.ack(data);
      await database_operations.index_webpages(db, deserialize_data);

      database_channel.sendToQueue(
        data.properties.replyTo,
        Buffer.from(
          JSON.stringify({
            isSuccess: deserialize_data.CrawlStatus,
            Message: deserialize_data.Message,
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
            isSuccess: false,
            Message:
              "Something went wrong with the database service, please retry the application.",
            Url: deserialize_data.Url,
            WebpageCount: 0,
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
      const data_query: Webpage[] =
        await database_operations.query_webpages(db);
      console.log({ searchEngineMessage: data.content.toString() });

      const MSS = 100000;
      let segments = segment_serializer.createSegments(data_query, MSS);
      console.log("Total segments created: %d", segments.length);

      // Need to find a way to get an ack notification from the message queue of
      // db_cbq_sengine so that we can send the next segment in the sequence
      database_channel.ack(data);
      segments.forEach(async (segment) => {
        await database_channel.sendToQueue(
          db_cbq_sengine, // respond back to this queue from search engine
          Buffer.from(segment),
          {
            correlationId: data.properties.correlationId,
          },
        );
      });
    } catch (err) {
      const error = err as Error;
      console.error(error.message);
      console.error(error.stack);
      database_channel.nack(data, false, false);
    }
  });
}

export default { channel_handler };
