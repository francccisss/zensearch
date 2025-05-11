import type { Channel, ConsumeMessage, ChannelModel } from "amqplib";
import amqp from "amqplib";
import EventEmitter from "stream";
import CircularBuffer from "../segments/circular_buffer.js";
import {
  DB_EXPRESS_CHECK_CBQ,
  CRAWLER_EXPRESS_CBQ,
  EXPRESS_CRAWLER_QUEUE,
  EXPRESS_DB_CHECK_QUEUE,
  EXPRESS_SENGINE_QUERY_QUEUE,
  SENGINE_EXPRESS_QUERY_CBQ,
} from "./routing_keys.js";
import type { CrawlMessageStatus } from "../types/index.js";

// TODO ADD LOGS TO RECEIVED AND PROCESSED SEGMENTS
class RabbitMQClient {
  connection: null | ChannelModel = null;
  client: this = this;
  searchChannel: Channel | null = null;
  crawlChannel: Channel | null = null;
  eventEmitter: EventEmitter = new EventEmitter();
  circleBuffer: CircularBuffer = new CircularBuffer(100);

  async establishConnection(retryCount: number): Promise<RabbitMQClient> {
    if (retryCount > 0) {
      retryCount--;
      try {
        this.connection = await amqp.connect("amqp://localhost:5672");
        console.log(
          `Successfully connected to rabbitmq after ${retryCount} retries`,
        );
        return this;
      } catch (err) {
        console.error("Retrying Web server service connection");
        await new Promise((resolve) => {
          const timeoutID = setTimeout(() => {
            resolve("Done blocking");
            clearTimeout(timeoutID);
          }, 2000);
        });
        return await this.establishConnection(retryCount);
      }
    }
    throw new Error("Shutting down web server after several retries");
  }

  async initChannelQueues() {
    try {
      if (this.connection == null) {
        throw new Error("ERROR: Connection interface is null.");
      }
      this.searchChannel = await this.connection.createChannel();
      this.crawlChannel = await this.connection.createChannel();

      this.searchChannel.assertQueue(SENGINE_EXPRESS_QUERY_CBQ, {
        exclusive: false,
        durable: false,
      });

      this.crawlChannel.assertQueue(CRAWLER_EXPRESS_CBQ, {
        exclusive: false,
        durable: false,
      });
    } catch (err) {
      const error = err as Error;
      console.error(
        "ERROR: Something went wrong while creating search channels.",
      );
      console.error(err);
      throw new Error(error.message);
    }
  }

  /*
   This Function sends a new search query to the SEARCH ENGINE SERVICE.
   returns a bool to check if it has been sent
  */
  async sendSearchQuery(q: string): Promise<boolean> {
    if (this.searchChannel == null) {
      throw new Error("ERROR: Search Channel is null");
    }
    return this.searchChannel.sendToQueue(
      EXPRESS_SENGINE_QUERY_QUEUE,
      Buffer.from(q),
      {
        replyTo: SENGINE_EXPRESS_QUERY_CBQ,
      },
    );
  }

  async crawlChannelListener(
    cb: (chan: Channel, data: ConsumeMessage, message_type: string) => void,
  ) {
    if (this.crawlChannel == null)
      throw new Error("ERROR: Crawl Channel is null.");

    try {
      this.crawlChannel.consume("", async (msg) => {
        if (msg === null) throw new Error("No Response");
        console.log("LOG: Message received from crawler");
        if (this.crawlChannel == null) {
          throw new Error("ERROR: Crawl Channel is null.");
        }
        const deserializedMessage = new TextDecoder().decode(msg.content);
        const crawlMessage: CrawlMessageStatus =
          JSON.parse(deserializedMessage);
        console.log(crawlMessage);
        //this.crawlChannel.ack(msg, true);
        cb(this.crawlChannel, msg, "crawling");
      });
    } catch (err) {
      const error = err as Error;
      console.log("LOG: Something went wrong with the channel listeners");
      console.error(error.name);
      console.error(error.message);
    }
  }

  // Channel Listener used by express server for listening for the
  // search channel queue callback for webpages retrieval

  async searchChannelListener() {
    try {
      if (this.searchChannel == null) {
        throw new Error("ERROR: Search channel does not exist.");
      }
      //this.searchChannel.assertQueue(SEARCH_ES_CB, {
      //  exclusive: false,
      //  durable: false,
      //});
      this.searchChannel.consume(
        SENGINE_EXPRESS_QUERY_CBQ,
        (data: ConsumeMessage | null) => {
          if (data === null) {
            this.searchChannel!.close();
            console.error("Msg does not exist");
            this.eventEmitter.emit("segmentError", {
              data: null,
              err: new Error(
                "Something went wrong while listening to segments",
              ),
            });
            return;
          }
          this.eventEmitter.emit("newSegment", { data, err: null });
        },
        { noAck: false },
      );
    } catch (err) {
      console.error(err);
    }
  }

  async addSegmentsToQueue() {
    this.eventEmitter.on("newSegment", (segment) => {
      this.circleBuffer.write(segment);
    });
  }

  /*
   Everytime a new segment arrives from the search engine the promise is resolved
   within the callback function of the event listener, this is synchronous
   in terms of data being perserved within the segmentGenerator.
  */

  async *segmentGenerator() {
    while (true) {
      if (this.circleBuffer.inUseSize() > 0) {
        yield this.circleBuffer.read();
      } else {
        await new Promise((resolve) => setImmediate(resolve));
      }
    }
  }

  // Crawler Expects an Object to be unmarshalled where the array of websites are inside a Docs property.
  async crawl(
    websites: Buffer,
    job: { queue: string; id: string },
  ): Promise<boolean> {
    if (this.connection === null)
      throw new Error("ERROR: TCP Connection lost.");
    const chan = await this.connection.createChannel();
    try {
      await chan.assertQueue(EXPRESS_CRAWLER_QUEUE, {
        exclusive: false,
        durable: false,
      });
      const success = chan.sendToQueue(EXPRESS_CRAWLER_QUEUE, websites, {
        replyTo: CRAWLER_EXPRESS_CBQ,
        correlationId: job.id,
      });
      if (!success) {
        throw new Error("ERROR: Unable to send to job to crawler service");
      }
      await chan.close();
      return true;
    } catch (err) {
      const error = err as Error;
      console.error("ERROR: Something went wrong while starting the crawl.");
      console.error(error.message);
      await chan.close();
      return false;
    }
  }

  /*
      Sends a message to the database service to check and see if the DOCS
      or list of websites the users want to crawl already exists in the database.
    */
  async crawlListCheck(
    encoded_list: Uint8Array<ArrayBufferLike>,
  ): Promise<null | {
    unindexed: Array<string>;
  }> {
    if (this.connection === null)
      throw new Error("Unable to create a channel for crawl queue.");
    const channel = await this.connection.createChannel();

    await channel.assertQueue(EXPRESS_DB_CHECK_QUEUE, {
      durable: false,
      exclusive: false,
    });
    await channel.assertQueue(DB_EXPRESS_CHECK_CBQ, {
      durable: false,
      exclusive: false,
    });

    channel.sendToQueue(EXPRESS_DB_CHECK_QUEUE, Buffer.from(encoded_list), {
      replyTo: DB_EXPRESS_CHECK_CBQ,
    });
    let unindexed: Array<string> = [];
    let isError = false;
    await channel.consume(DB_EXPRESS_CHECK_CBQ, async (data) => {
      if (data === null) {
        throw new Error("ERROR: Data received is null.");
      }
      try {
        const parseList: { Docs: Array<string> } = JSON.parse(
          data.content.toString(),
        );
        unindexed = parseList.Docs;
        channel.ack(data);
      } catch (err) {
        isError = true;
        channel.nack(data, false, false);
      }
    });

    isError ??
      console.error(
        "ERROR: Something went wrong while tring to listen to DB serveice",
      );

    channel.close();
    return isError ? null : { unindexed };
  }
}

// Start a new tcp connection with the rabbitmq server
const client = new RabbitMQClient();
export default { client };
