import amqp from "amqplib";
import databaseOperations from "../database";
import { Database } from "sqlite3";
import { IndexedWebpages, Webpage } from "../utils/types";
import segmentSerializer from "../serializer/segment_serializer";
import {
  CRAWLER_DB_INDEXING_NOTIF_QUEUE,
  DB_EXPRESS_CHECK_CBQ,
  DB_CRAWLER_INDEXING_NOTIF_CBQ,
  DB_SENGINE_REQUEST_CBQ,
  EXPRESS_DB_CHECK_QUEUE,
  SENGINE_DB_REQUEST_QUEUE,
} from "./routing_keys";

/*
  channel handler can take in multiple channels from a single tcp conneciton
  to the rabbitmq message broker, these channels are multiplexed to handle
  messages coming from different context eg: database and search
*/

async function channelHandler(db: Database, databaseChannel: amqp.Channel) {
  // EXPRESS SERVER ROUTING KEYS
  // routing key used by express server to check existing webpages.

  /*
   TODO Document code please :)
   TODO Change names to make it more comprehensible please :D
   TODO Use PRE-Compression eg:
    saving new indexed webpage should be compressed and decompressed on
    request

    Consumer waits for the crawler service to push new webpages into the `db_indexing_crawler`
    message queue, the `indexWebpages` handler saves these crawled webpages into
    the database.
  */
  await databaseChannel.assertQueue(CRAWLER_DB_INDEXING_NOTIF_QUEUE, {
    exclusive: false,
    durable: false,
  });

  // Errors and successful crawls will be demultiplexed here
  databaseChannel.consume(CRAWLER_DB_INDEXING_NOTIF_QUEUE, async (data) => {
    // who is going to catch this error? aaaaaa
    if (data === null) throw new Error("No data was pushed.");
    const decoder = new TextDecoder();
    const decodedData = decoder.decode(data.content as ArrayBuffer);
    const deserializeData: IndexedWebpages = JSON.parse(decodedData);
    try {
      databaseChannel.ack(data);
      //await databaseOperations.indexWebpages(db, deserializeData);
      console.log("Storing data");
      databaseChannel.sendToQueue(
        DB_CRAWLER_INDEXING_NOTIF_CBQ,
        Buffer.from(
          JSON.stringify({
            isSuccess: true,
            Message: "Successfully stored webpages",
            URLSeed: deserializeData.URLSeed,
          }),
        ),
      );
    } catch (err) {
      const error = err as Error;
      console.error("ERROR: %s", error.message);
      console.error("ERROR: %s", deserializeData.Message);
      console.log("Sending back response to crawler");
      databaseChannel.sendToQueue(
        DB_CRAWLER_INDEXING_NOTIF_CBQ,
        Buffer.from(
          JSON.stringify({
            IsSuccess: false,
            Message: error.message,
            URLSeed: deserializeData.URLSeed,
          }),
        ),
      );
    }
  });

  // SEARCH ENGINE AND EXPRESS JS CONSUMERS
  await databaseChannel.assertQueue(SENGINE_DB_REQUEST_QUEUE, {
    exclusive: false,
    durable: false,
  });

  await databaseChannel.assertQueue(EXPRESS_DB_CHECK_QUEUE, {
    exclusive: false,
    durable: false,
  });

  /*
    The `ES_DB_CHECK_QUEUE` routing key used to consumes the messages from the express server
    for new crawl tasks, its responsibility is to check send
    the array of crawl tasks from the client to the database service
    and query the database to see if the list of crawl tasks exists
    on the database already by checking the indexed_sites TABLE,
    if it does it returns back the array filtering out the ones that
    have already been crawled and indexed into the database, returning
    the remaining items as uncrawled to be processed by the crawler service.
   */

  databaseChannel.consume(EXPRESS_DB_CHECK_QUEUE, async (data) => {
    if (data == null) throw new Error("No data was pushed.");
    try {
      console.log(
        "NOTIF: DB service received crawl list to check existing indexed websites.",
      );
      const crawlList: { Docs: Array<string> } = JSON.parse(
        data.content.toString(),
      );
      const unindexedWebsites = await databaseOperations.checkExistingTasks(
        db,
        crawlList.Docs,
      );

      const encoder = new TextEncoder();
      const encodedDocs = encoder.encode(
        JSON.stringify({ Docs: unindexedWebsites }),
      );

      databaseChannel.ack(data);
      const is_sent = databaseChannel.sendToQueue(
        DB_EXPRESS_CHECK_CBQ,
        Buffer.from(encodedDocs),
      );
      if (!is_sent) {
        throw new Error("ERROR: Unable to send back message.");
      }
    } catch (err) {
      const error = err as Error;
      databaseChannel.nack(data, false, false);
      console.error(error.message);
      console.error(err);
    }
  });

  /*
    Consumes messages sent by the search engine services to query the database for
    the webpages to be ranked and sent to the client.

    a callback queue is assigned once we consume and query webpages so that we can send
    it back right after all those process are done.
  */
  databaseChannel.consume(SENGINE_DB_REQUEST_QUEUE, async (data) => {
    if (data === null) throw new Error("No data was pushed.");
    try {
      const dataQuery: Webpage[] = await databaseOperations.queryWebpages(db);
      console.log({ searchEngineMessage: data.content.toString() });

      databaseChannel.ack(data);
      const MSS = 100000;
      let segments = segmentSerializer.createSegments(
        Buffer.from(JSON.stringify(dataQuery)),
        MSS,
        async (newSegment: Buffer) => {
          databaseChannel.sendToQueue(
            DB_SENGINE_REQUEST_CBQ, // respond back to this queue from search engine
            newSegment,
            {
              correlationId: data.properties.correlationId,
            },
          );
        },
      );
      console.log("Total segments created: %d", segments.length);
    } catch (err) {
      const error = err as Error;
      console.error(error.message);
      console.error(error.stack);
      databaseChannel.nack(data, false, false);
    }
  });
}

export default { channelHandler };
