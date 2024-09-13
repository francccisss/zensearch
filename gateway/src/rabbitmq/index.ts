import amqp, { Connection, Channel } from "amqplib";
import {
  CRAWL_QUEUE,
  CRAWL_QUEUE_CB,
  SEARCH_QUEUE,
  SEARCH_QUEUE_CB,
} from "./queues";
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

async function search_job(
  job: { q: string; job_id: string },
  conn: Connection,
) {
  const channel = await conn.createChannel();
  await channel.assertQueue(SEARCH_QUEUE, {
    exclusive: false,
    durable: false,
  });

  await channel.assertQueue(SEARCH_QUEUE_CB, {
    exclusive: false,
    durable: false,
  });
  const success = await channel.sendToQueue(SEARCH_QUEUE, Buffer.from(job.q), {
    correlationId: job.job_id,
    replyTo: SEARCH_QUEUE_CB,
  });

  await channel.consume(SEARCH_QUEUE_CB, async (msg) => {
    if (msg === null) throw new Error("Msg does not exist");
    console.log(msg.toString());
    if (msg.properties.correlationId === job.job_id) {
      return msg.content.toString();
    }
  });

  if (!success) {
    await channel.close();
    throw new Error("Unable to send job to search queue.");
  }
  await channel.close();
}

async function poll_job(
  chan: Channel,
  job: { id: string; queue: string },
): Promise<{ done: boolean; data: any }> {
  try {
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
        for (let page of JSON.parse(data)) {
          console.log(page.Title);
          console.log({ TFIDFRating: page.TFIDFRating, TFScore: page.TFScore });
        }
        console.log("CONSUMED");
        chan.ack(response);
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
