import amqp from "amqplib";
import sql from "../database.js";
import Database from "better-sqlite3";
import type {
  DequeuedUrl,
  IndexedWebpage,
  URLs,
  Webpage,
} from "../utils/types.js";
import segmentSerializer from "../serializer/segment_serializer.js";
import {
  CRAWLER_DB_INDEXING_QUEUE,
  DB_SENGINE_REQUEST_CBQ,
  EXPRESS_DB_CHECK_QUEUE,
  SENGINE_DB_REQUEST_QUEUE,
} from "./routing_keys.js";
import database from "../database.js";

export async function establishConnection(
  retries: number,
): Promise<amqp.ChannelModel> {
  console.log("Notif: Connecting...");
  if (retries > 0) {
    retries--;
    try {
      const connection = await amqp.connect("amqp://localhost:5672");
      console.log(`Successfully connected to rabbitmq after several retries`);
      return connection;
    } catch (err) {
      console.error("Retrying Database service connection...");
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
/*
  channel handler can take in multiple channels from a single tcp conneciton
  to the rabbitmq message broker, these channels are multiplexed to handle
  messages coming from different context eg: database and search
*/

async function webpageHandler(
  db: Database.Database,
  databaseChannel: amqp.Channel,
) {
  // EXPRESS SERVER ROUTING KEYS
  // routing key used by express server to check existing webpages.

  /*
   TODO Document code please :)
   TODO Use PRE-Compression eg:
    saving new indexed webpage should be compressed and decompressed on
    request

    Consumer waits for the crawler service to push new webpages into the `db_indexing_crawler`
    message queue, the `indexWebpages` handler saves these crawled webpages into
    the database.
  */
  await databaseChannel.assertQueue(CRAWLER_DB_INDEXING_QUEUE, {
    exclusive: false,
    durable: false,
  });

  // Errors and successful crawls will be demultiplexed here
  databaseChannel.consume(CRAWLER_DB_INDEXING_QUEUE, async (data) => {
    // who is going to catch this error? aaaaaa
    if (data === null) throw new Error("No data was pushed.");
    const decoder = new TextDecoder();
    const decodedData = decoder.decode(data.content as unknown as ArrayBuffer);
    const deserializeData: IndexedWebpage = JSON.parse(decodedData);
    try {
      databaseChannel.ack(data);
      database.saveWebpage(db, deserializeData);
      console.log("Storing data");
      databaseChannel.sendToQueue(
        data.properties.replyTo,
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
      console.error("ERROR: %s", error);
      console.log("Sending back response to crawler");
      console.log(deserializeData);
      databaseChannel.sendToQueue(
        data.properties.replyTo,
        Buffer.from(
          JSON.stringify({
            IsSuccess: false,
            Message: "Unable to store indexed webpages to sqlite",
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
      const unindexedWebsites = sql.checkAlreadyIndexedWebpage(
        db,
        crawlList.Docs,
      );

      const encoder = new TextEncoder();
      const encodedDocs = encoder.encode(
        JSON.stringify({ Docs: unindexedWebsites }),
      );

      databaseChannel.ack(data);
      const is_sent = databaseChannel.sendToQueue(
        data.properties.replyTo,
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
      const dataQuery: Webpage[] = await sql.queryWebpages(db);
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

async function frontierQueueHandler(
  db: Database.Database,
  frontierChannel: amqp.Channel,
) {
  const CRAWLER_DB_DEQUEUE_URL_QUEUE = "crawler_db_dequeue_url_queue";

  const CRAWLER_DB_STOREURLS_QUEUE = "crawler_db_storeurls_queue";
  const CRAWLER_DB_CLEARURLS_QUEUE = "crawler_db_clearurls_queue";

  const CRAWLER_DB_GET_LEN_QUEUE = "crawler_db_len_queue";

  await frontierChannel.assertQueue(CRAWLER_DB_DEQUEUE_URL_QUEUE, {
    exclusive: false,
    durable: false,
  });

  await frontierChannel.assertQueue(CRAWLER_DB_GET_LEN_QUEUE, {
    exclusive: false,
    durable: false,
  });

  await frontierChannel.assertQueue(CRAWLER_DB_STOREURLS_QUEUE, {
    exclusive: false,
    durable: false,
  });

  await frontierChannel.assertQueue(CRAWLER_DB_CLEARURLS_QUEUE, {
    exclusive: false,
    durable: false,
  });

  frontierChannel.consume(
    CRAWLER_DB_DEQUEUE_URL_QUEUE,
    async (msg: amqp.ConsumeMessage | null) => {
      try {
        if (msg == null) {
          throw new Error("Message is null");
        }

        frontierChannel.ack(msg);
        console.log("DATABASE TEST: DEQUEUEING");
        const domain = msg.content.toString();
        const { length, url, inProgressNode, message } = database.dequeueURL(
          db,
          domain,
        );

        const dequeuedUrl: DequeuedUrl = { RemainingInQueue: length, Url: url };
        const msgBuffer = Buffer.from(JSON.stringify(dequeuedUrl));

        const sent = frontierChannel.sendToQueue(
          msg.properties.replyTo,
          msgBuffer,
        );
        if (!sent) {
          throw new Error("Error: Unable to send a dequeueded URL");
        }

        console.log("TEST DATABASE: DEQUEUED NODE ", dequeuedUrl);
        console.log("TEST DATABASE: DEQUEUE MESSAGE: %s", message);
        console.log(inProgressNode);
        // node can be null if queue is empty
        if (inProgressNode !== null) {
          database.setNodeToVisited(db, inProgressNode);
          console.log("Node updated to visited, remove in_progress node.");
        }

        if (inProgressNode == null && length == 0) {
          console.log("DATABASE TEST: REMOVED QUEUE");
          database.removeQueue(db, domain);
        }
      } catch (err) {
        console.log(err);
        // i dont know what to do with this yet
      }
    },
  );
  frontierChannel.consume(
    CRAWLER_DB_STOREURLS_QUEUE,
    (msg: amqp.ConsumeMessage | null) => {
      try {
        if (msg == null) {
          throw new Error("Message is null");
        }
        const URLs: URLs = JSON.parse(msg.content.toString());
        console.log("DATABASE TEST: URLS ", URLs);
        database.enqueueUrls(db, URLs);
        frontierChannel.ack(msg);
      } catch (err) {
        // i dont know what to do with this yet
        console.log(err);
      }
    },
  );

  frontierChannel.consume(
    CRAWLER_DB_GET_LEN_QUEUE,
    (msg: amqp.ConsumeMessage | null) => {
      try {
        if (msg == null) {
          throw new Error("Message is null");
        }
        frontierChannel.ack(msg);
        const hostname = msg.content.toString();
        const queueLen = database.getCurrentQueueLen(db, hostname);

        const queueLenBuf = Buffer.alloc(4);
        queueLenBuf.writeIntLE(queueLen, 0, 4);
        console.log("DATABASE TEST: QUEUE LEN BUFFER ", queueLenBuf);
        console.log("DATABASE TEST: QUEUE='%s' LENGTH ", hostname, queueLen);

        frontierChannel.sendToQueue(msg.properties.replyTo, queueLenBuf);
      } catch (err) {
        // i dont know what to do with this yet
        console.log(err);
      }
    },
  );
}
export default { establishConnection, webpageHandler, frontierQueueHandler };
