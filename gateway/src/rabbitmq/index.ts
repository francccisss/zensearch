import amqp, { Connection, Channel, ConsumeMessage } from "amqplib";
import {
  CRAWL_QUEUE,
  CRAWL_QUEUE_CB,
  SEARCH_QUEUE,
  SEARCH_QUEUE_CB,
} from "./routing_keys";

class RabbitMQClient {
  connection: null | Connection = null;
  client: this = this;
  search_channel: Channel | null = null;

  constructor() {}

  async connectClient() {
    /*
     Check if there already exists a connection if not create a new connection else just return
     A single pattern to make sure that we are only creating a single tcp connection
    */
    if (this.connection == null) {
      try {
        this.connection = await amqp.connect("amqp://localhost");
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

  async init_search_channel_queues() {
    try {
      if (this.connection == null) {
        throw new Error("ERROR: Connection interface is null.");
      }
      this.search_channel = await this.connection.createChannel();
      await this.search_channel.assertQueue(SEARCH_QUEUE, {
        exclusive: false,
        durable: false,
      });
      await this.search_channel.assertQueue(SEARCH_QUEUE_CB, {
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
*/
  async send_search_query(job: {
    q: string;
    job_id: string;
  }): Promise<boolean> {
    if (this.search_channel == null) {
      return false;
    }
    return await this.search_channel.sendToQueue(
      SEARCH_QUEUE,
      Buffer.from(job.q),
      {
        correlationId: job.job_id,
        replyTo: SEARCH_QUEUE_CB,
      },
    );
  }

  /*
  A listener for consuming search engine's search results that is pushed to the message queue
  using `CRAWL_QUEUE_CB` routing key after it finishes processing user's search query.

  Takes in a callback function argument to process the data received from
  the search engine service.

*/
  async search_channel_listener(cb: (data: ConsumeMessage | null) => void) {
    try {
      if (this.search_channel == null)
        throw new Error("ERROR: Search Channel is null.");
      await this.search_channel.consume(
        SEARCH_QUEUE_CB,
        // errors will propogate outside
        async (msg: ConsumeMessage | null) => {
          if (msg === null) throw new Error("Msg does not exist");
          console.log(msg);
          cb(msg);
          //I hate this
          if (this.search_channel == null) {
            throw new Error("ERROR: Search Channel is null.");
          }
          this.search_channel.ack(msg);
        },
      );
    } catch (err) {
      // Propogate error
      const error = err as Error;
      throw new Error(error.message);
    }
  }

  /*
   * These methods underneath are used in the express server
   * their purpose is to push messages into the message queue for every user
   * request that is sent to the server.
   * `poll_job()` function is used to poll specific job using the job_id identifier that
   * was set by the server in the client's cookies
   */
  async poll_job(job: {
    id: string;
    queue: string;
  }): Promise<{ done: boolean; data: any }> {
    try {
      if (this.connection === null) throw new Error("TCP Connection lost.");
      const chan = await this.connection.createChannel();
      await chan.assertQueue(job.queue as string, {
        exclusive: false,
        durable: false,
      });
      const { queue, messageCount, consumerCount } = await chan.checkQueue(
        job.queue as string,
      );
      if (messageCount === 0) {
        return { done: false, data: {} };
      }
      let data: any;
      const consumer = await chan.consume(
        job.queue as string,
        async (response) => {
          if (response === null) throw new Error("No Response");
          console.log("LOG: Response from Polled Job received");
          data = response.content.toString();
          console.log("CONSUMED");
          chan.ack(response);
        },
      );
      return { done: true, data };
    } catch (err) {
      const error = err as Error;
      console.error("ERROR: Something went wrong while polling message queue");
      console.error(error.message);
      throw new Error(error.message);
    }
  }

  async crawl(
    websites: Uint8Array,
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

      await chan.assertQueue(CRAWL_QUEUE_CB, {
        exclusive: false,
        durable: false,
      });
      const success = await chan.sendToQueue(
        CRAWL_QUEUE,
        Buffer.from(websites.buffer),
        {
          replyTo: CRAWL_QUEUE_CB,
          correlationId: job.id,
        },
      );
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

  // TODO handle errors in here please :D

  async crawl_list_check(encoded_list: ArrayBuffer): Promise<null | {
    undindexed: Array<string>;
    data_buffer: Buffer;
  }> {
    if (this.connection === null)
      throw new Error("Unable to create a channel for crawl queue.");
    const channel = await this.connection.createChannel();

    /*
      Sends a message to the database service to check and see if the DOCS
      or list of websites the users want to crawl already exists in the database.
    */

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
