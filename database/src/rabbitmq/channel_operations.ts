import amqp from "amqplib";
import databaseOperations from "../database_operations";
import { Database } from "sqlite3";
import { Message, Webpage } from "../utils/types";
import segmentSerializer from "../serializer/segment_serializer";

/*
  channel handler can take in multiple channels from a single tcp conneciton
  to the rabbitmq message broker, these channels are multiplexed to handle
  messages coming from different context eg: database and search
*/

async function channelHandler(db: Database, databaseChannel: amqp.Channel) {
  // CRAWLER ROUTING KEYS
  const DB_INDEXING_CRAWLER = "db_indexing_crawler";

  // SEARCH ENGINE ROUTING KEYS
  // routing key used by search engine service to query database for webpages.
  const DB_QUERY_SENGINE = "db_query_sengine";
  // routing key to reply back to the search engine service's callback queue.
  const DB_CBQ_SENGINE = "db_cbq_sengine";

  // EXPRESS SERVER ROUTING KEYS
  // routing key used by express server to check existing webpages.
  const DB_CHECK_EXPRESS = "db_check_express";
  const DB_CBQ_EXPRESS = "db_cbq_express";
  const DB_CBQ_POLL_EXPRESS = "crawl_poll_queue";

  /*
   TODO Document code please :)
   TODO Change names to make it more comprehensible please :D

    Consumer waits for the crawler service to push new webpages into the `db_indexing_crawler`
    message queue, the `indexWebpages` handler saves these crawled webpages into
    the database.
  */
  await databaseChannel.assertQueue(DB_INDEXING_CRAWLER, {
    exclusive: false,
    durable: false,
  });

  // Aside from database service to handle errors for indexing,
  // the crawler also needs to send the same mssage format straight
  // to the express server to notify that the crawl for the current url
  // threw an error.

  databaseChannel.consume(DB_INDEXING_CRAWLER, async (data) => {
    if (data === null) throw new Error("No data was pushed.");
    const decoder = new TextDecoder();
    const decodedData = decoder.decode(data.content as ArrayBuffer);
    const deserializeData: Message = JSON.parse(decodedData);
    try {
      databaseChannel.ack(data);
      await databaseOperations.indexWebpages(db, deserializeData);
      databaseChannel.sendToQueue(
        data.properties.replyTo,
        Buffer.from(
          JSON.stringify({
            isSuccess: deserializeData.CrawlStatus,
            Message: deserializeData.Message,
            Url: deserializeData.Url,
            WebpageCount: deserializeData.Webpages.length,
          }),
        ),
      );
    } catch (err) {
      const error = err as Error;
      console.error("ERROR: Decoder was unable to deserialized indexed data.");
      console.error(error.message);
      databaseChannel.nack(data, false, false);
      databaseChannel.sendToQueue(
        data.properties.replyTo,
        Buffer.from(
          JSON.stringify({
            isSuccess: false,
            Message:
              "Something went wrong with the database service, please retry the application.",
            Url: deserializeData.Url,
            WebpageCount: 0,
          }),
        ),
      );
    }
  });

  // SEARCH ENGINE AND EXPRESS JS CONSUMERS
  await databaseChannel.assertQueue(DB_QUERY_SENGINE, {
    exclusive: false,
    durable: false,
  });

  await databaseChannel.assertQueue(DB_CHECK_EXPRESS, {
    exclusive: false,
    durable: false,
  });

  /*
    The `db_check_express`consumes the messages from the express server
    for new crawl tasks, its responsibility is to check send
    the array of crawl tasks from the client to the database service
    and query the database to see if the list of crawl tasks exists
    on the database already by checking the indexed_sites TABLE,
    if it does it returns back the array filtering out the ones that
    have already been crawled and indexed into the database, returning
    the remaining items as uncrawled to be processed by the crawler service.
   */

  databaseChannel.consume(DB_CHECK_EXPRESS, async (data) => {
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
        DB_CBQ_EXPRESS,
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

    a callback queue is implemented once we consume and query webpages so that we can send
    it back right after all those process are done.
  */
  databaseChannel.consume(DB_QUERY_SENGINE, async (data) => {
    if (data === null) throw new Error("No data was pushed.");
    try {
      const data_query: Webpage[] = await databaseOperations.queryWebpages(db);
      console.log({ searchEngineMessage: data.content.toString() });
      const MSS = 100000;
      let segments = segmentSerializer.createSegments(data_query, MSS);
      console.log("Total segments created: %d", segments.length);

      // Need to find a way to get an ack notification from the message queue of
      // db_cbq_sengine so that we can send the next segment in the sequence
      databaseChannel.ack(data);
      segments.forEach(async (segment) => {
        await databaseChannel.sendToQueue(
          DB_CBQ_SENGINE, // respond back to this queue from search engine
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
      databaseChannel.nack(data, false, false);
    }
  });
}

export default { channelHandler };
