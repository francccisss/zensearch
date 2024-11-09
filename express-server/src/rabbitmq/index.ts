import amqp, { Connection, Channel, ConsumeMessage } from "amqplib";
import {
  CRAWL_QUEUE,
  CRAWL_QUEUE_CB,
  SEARCH_QUEUE,
  SEARCH_QUEUE_CB,
} from "./routing_keys";
import { EventEmitter } from "stream";

class RabbitMQClient {
  connection: null | Connection = null;
  client: this = this;
  search_channel: Channel | null = null;
  crawl_channel: Channel | null = null;
  crawledSites: Array<string> = []; // string as an utf-8 buffer
  eventEmitter: EventEmitter = new EventEmitter();

  constructor() {}

  async connectClient() {
    /*
     Check if there already exists a connection if not create a new connection else just return
     A single pattern to make sure that we are only creating a single tcp connection
    */
    if (this.connection == null) {
      try {
        this.connection = await amqp.connect("amqp://rabbitmq:5672");
      } catch (err) {
        const error = err as Error;
        console.error(
          "ERROR:Unable establish a tcp connection with rabbitmq server.",
        );

        console.error(error);
        throw new Error(error.message);
      }
    }
    return this;
  }

  async init_channel_queues() {
    try {
      if (this.connection == null) {
        throw new Error("ERROR: Connection interface is null.");
      }
      this.search_channel = await this.connection.createChannel();
      this.crawl_channel = await this.connection.createChannel();

      this.search_channel.assertQueue(SEARCH_QUEUE, {
        exclusive: false,
        durable: false,
      });

      this.crawl_channel.assertQueue(CRAWL_QUEUE_CB, {
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
  async send_search_query(q: string): Promise<boolean> {
    if (this.search_channel == null) {
      throw new Error("ERROR: Search Channel is null");
    }
    return await this.search_channel.sendToQueue(SEARCH_QUEUE, Buffer.from(q), {
      replyTo: SEARCH_QUEUE_CB,
    });
  }

  /*
   A listener for consuming search engine's search results that is pushed to the message queue
   using `SEARCH_QUEUE_CB` routing key after it finishes processing user's search query.

   Takes in a callback function argument to process the data received from
   the search engine service.
  */
  async websocket_channel_listener(
    cb: (chan: Channel, data: ConsumeMessage, message_type: string) => void,
  ) {
    if (this.crawl_channel == null)
      throw new Error("ERROR: Crawl Channel is null.");

    try {
      // Consumes database service's output that was sent by the crawler service.
      // client -> ws(CRAWL_QUEUE_CB)[CRAWL_QUEUE] -> [CRAWL_QUEUE]Crawler(CRAWL_QUEUE_CB)[db_indexing_crawler]
      // -> [db_indexing_crawler]Database[CRAWL_QUEUE_CB] -> [CRAWL_QUEUE_CB]ws this listener -> client.
      // Crawler service directs database service to send a message to the message queue with CRAWL_QUEUE_CB
      // routing key after it finishes storing the indexed websites
      this.crawl_channel.consume(CRAWL_QUEUE_CB, async (msg) => {
        if (msg === null) throw new Error("No Response");
        console.log("LOG: Message received from crawling");
        if (this.crawl_channel == null) {
          throw new Error("ERROR: Crawl Channel is null.");
        }
        await cb(this.crawl_channel, msg, "crawling");
        console.log(msg.content.toString());
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

  /*
   * TODO retrieve data segments from search engine using the same method
   * used in the search engine by when retrieving from database.
   */
  async search_channel_listener() {
    try {
      if (this.search_channel == null) {
        throw new Error("ERROR: Search channel does not exist.");
      }
      this.search_channel.assertQueue(SEARCH_QUEUE_CB, {
        exclusive: false,
        durable: false,
      });
      await this.search_channel.consume(
        SEARCH_QUEUE_CB,
        (data: ConsumeMessage | null) => {
          if (data === null) {
            this.search_channel!.close();
            throw new Error("Msg does not exist");
          }
          // bruh it should already exist if we're calling consume.. the frick!
          this.search_channel!.ack(data);
          this.eventEmitter.emit("searchResults", { data, err: null });
        },
        { noAck: false },
      );
    } catch (err) {
      this.eventEmitter.emit("searchResults", { data: null, err });
      console.error(err);
    }
  }

  // Crawler Expects an Object to be unmarshalled where the
  // Array of websites are inside a Docs property.
  async crawl(
    websites: Buffer,
    job: { queue: string; id: string },
  ): Promise<boolean> {
    if (this.connection === null)
      throw new Error("ERROR: TCP Connection lost.");
    const chan = await this.connection.createChannel();
    const message = "Start Crawl";
    try {
      await chan.assertQueue(CRAWL_QUEUE, {
        exclusive: false,
        durable: false,
      });
      const success = await chan.sendToQueue(CRAWL_QUEUE, websites, {
        replyTo: CRAWL_QUEUE_CB,
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
  async crawl_list_check(encoded_list: ArrayBuffer): Promise<null | {
    undindexed: Array<string>;
    data_buffer: Buffer;
  }> {
    if (this.connection === null)
      throw new Error("Unable to create a channel for crawl queue.");
    const channel = await this.connection.createChannel();

    const db_check_queue = "db_check_express";
    const db_check_response_queue = "db_cbq_express";
    await channel.assertQueue(db_check_queue, {
      durable: false,
      exclusive: false,
    });
    await channel.assertQueue(db_check_response_queue, {
      durable: false,
      exclusive: false,
    });

    await channel.sendToQueue(db_check_queue, Buffer.from(encoded_list));
    let undindexed: Array<string> = [];
    let data_buffer: Buffer = Buffer.from("");
    let is_error = false;
    await channel.consume(db_check_response_queue, async (data) => {
      if (data === null) {
        throw new Error("ERROR: Data received is null.");
      }
      try {
        const parse_list: { Docs: Array<string> } = JSON.parse(
          data.content.toString(),
        );
        undindexed = parse_list.Docs;
        data_buffer = data.content;
        channel.ack(data);
      } catch (err) {
        is_error = true;
        channel.nack(data, false, false);
      }
    });

    channel.close();
    return is_error ? null : { undindexed, data_buffer };
  }
}

// Start a new tcp connection with the rabbitmq server
const client = new RabbitMQClient();
export default { client };
