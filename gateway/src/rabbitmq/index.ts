import amqp, { Connection, Channel, ConsumeMessage } from "amqplib";
import {
  CRAWL_QUEUE,
  CRAWL_QUEUE_CB,
  SEARCH_QUEUE,
  SEARCH_QUEUE_CB,
} from "./routing_keys";
let connection: Connection | null = null;
let search_channel: Channel;

class RabbitMQClient {
  connection: null | Connection = null;
  client: this = this;

  async connectClient() {
    /*
     Check if there already exists a connection if not create a new connection else just return
     A single pattern to make sure that we are only creating a single tcp connection
    */
    if (this.connection == null) {
      this.connection = await amqp.connect("amqp://localhost");
    }
    return this;
  }

  async init_search_channel_queues() {
    try {
      if (connection == null) {
        throw new Error(
          "Unable establish a tcp connection with rabbitmq server.",
        );
      }
      search_channel = await connection.createChannel();
      await search_channel.assertQueue(SEARCH_QUEUE, {
        exclusive: false,
        durable: false,
      });
      await search_channel.assertQueue(SEARCH_QUEUE_CB, {
        exclusive: false,
        durable: false,
      });
    } catch (err) {
      console.error(
        "ERROR: Something went wrong while creating search channels.",
      );
      console.error(err);
    }
  }

  /*
  This Function sends a new search query to the SEARCH ENGINE SERVICE.
  TODO error handle this please. :D
*/
  async send_search_query(job: {
    q: string;
    job_id: string;
  }): Promise<boolean> {
    return await search_channel.sendToQueue(SEARCH_QUEUE, Buffer.from(job.q), {
      correlationId: job.job_id,
      replyTo: SEARCH_QUEUE_CB,
    });
  }

  /*
  A listener for consuming search engine's search results that is pushed to the message queue
  using `CRAWL_QUEUE_CB` routing key after it finishes processing user's search query.

  Takes in a callback function argument to process the data received from
  the search engine service.

  TODO Handle errors in here please :D
*/
  async search_channel_listener(cb: (data: ConsumeMessage | null) => void) {
    await search_channel.consume(
      SEARCH_QUEUE_CB,
      async (msg: ConsumeMessage | null) => {
        if (msg === null) throw new Error("Msg does not exist");
        console.log(msg);
        cb(msg);
        search_channel.ack(msg);
      },
    );
  }

  /*
   * These methods underneath are used in the express server
   * their purpose is to push messages into the message queue for every user
   * request that is sent to the server.
   *
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
      console.log(queue);
      console.log(messageCount);
      console.log(consumerCount);
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
          return { done: true, data };
        },
      );
      return { done: true, data };
    } catch (err) {
      const error = err as Error;
      console.log("LOG:Something went wrong while polling queue");
      console.error(error.message);
      throw new Error(error.message);
    }
  }

  async crawl_job(websites: Uint8Array, job: { queue: string; id: string }) {
    if (this.connection === null) throw new Error("TCP Connection lost.");
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
        throw new Error("Unable to send to job to crawl queue");
      }
      await chan.close();
      return true;
    } catch (err) {
      const error = err as Error;
      console.error("LOG: Something went wrong while starting the crawl.");
      console.error(error.message);
      await chan.close();
      return false;
    }
  }
}

// Start a new tcp connection with the rabbitmq server
const client = new RabbitMQClient();
export default { client };
