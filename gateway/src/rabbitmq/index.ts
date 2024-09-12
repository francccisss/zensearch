import amqp, { Connection, Channel } from "amqplib";
import { CRAWL_QUEUE, CRAWL_QUEUE_CB } from "./queues";
let connection: Connection | null = null;

async function connect(): Promise<Connection | null> {
  if (connection) {
    return connection;
  }
  try {
    connection = await amqp.connect("amqp://localhost");
    console.log("Connected to RabbitMQ");
    return connection;
  } catch (error) {
    console.error("Failed to connect to RabbitMQ:", error);
    throw error;
  }
}

async function search_job(search_query: string, conn: Connection) {
  const channel = await conn.createChannel();
  const queue = "search_queue";
  const rps_queue = "search_rps_queue";
  const cor_id = "a29a5dec-fd24-4db4-83f1-db6dbefdaa6b";
  await channel.assertQueue(queue, {
    exclusive: false,
    durable: false,
  });
  const success = await channel.sendToQueue(queue, Buffer.from(search_query));
  if (!success) {
    throw new Error("Unable to send job to search queue.");
  }
  await channel.close();
}

async function poll_job(
  chan: Channel,
  job: { id: string; queue: string },
): Promise<boolean> {
  try {
    await chan.assertQueue(job.queue as string, {
      exclusive: false,
      durable: false,
    });
    const { queue, messageCount, consumerCount } = await chan.checkQueue(
      job.queue as string,
    );

    // TODO INDEFINITE PROCESSING PHASE IF CRAWLING SERVICE CRASHES
    // POLLING EVENT WILL KEEP RUNNING NEED TO CHECK IF SERVICE IS DOWN

    if (messageCount === 0) {
      return false;
    }
    const consumer = await chan.consume(
      job.queue as string,
      async (response) => {
        if (response === null) throw new Error("No Response");
        if (response.properties.correlationId === job.id) {
          console.log(
            "LOG: Response from Polled Job received: %s",
            response.content.toString(),
          );
          console.log("CONSUMED");
          await chan.close();
        }
      },
      { noAck: true },
    );
    return true;
  } catch (err) {
    const error = err as Error;
    console.log("LOG:Something went wrong while polling queue");
    console.error(error.message);
    throw new Error(error.message);
  }
}

async function crawl_job(
  chan: Channel,
  websites: Uint8Array,
  job: { queue: string; id: string },
) {
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

export default { connect, search_job, poll_job, crawl_job };
