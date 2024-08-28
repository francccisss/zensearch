import { EventEmitter } from "stream";
import utils from "./utils";
import WorkerHandler from "./WorkerHandler";
import path from "path";
import amqp, { Channel, Connection } from "amqplib";

const event = new EventEmitter();
const MAX_THREADS = 2;
console.log("Crawl start.");

(async function () {
  const connection = await amqp.connect("amqp://localhost");
  const channel = await connection.createChannel();
  var queue = "crawl_rpc_queue";
  await channel.assertQueue(queue, {
    durable: false,
  });
  console.log(" [*] Waiting for messages in %s. To exit press CTRL+C", queue);
  const decoder = new TextDecoder();

  channel.consume(
    queue,
    async function (msg) {
      if (msg === null) throw new Error("No message");
      const decoded_array_buffer = decoder.decode(msg.content);
      const to_json: { docs: Array<string> } = JSON.parse(decoded_array_buffer);
      console.log(to_json);
      const handler = await crawl(to_json.docs);
      let available_threads = handler.current_threads;
      await thread_polling(
        available_threads,
        handler.check_threads.bind(handler),
      );
      channel.sendToQueue(
        msg.properties.replyTo,
        Buffer.from(
          `Crawled: ${handler.successful_thread_count}/${to_json.docs.length}`,
        ),
        {
          correlationId: msg.properties.correlationId,
        },
      );
    },
    {
      noAck: true,
    },
  );
})();

async function thread_polling(
  available_threads: number,
  check_current_threads: () => number,
) {
  while (available_threads < MAX_THREADS) {
    available_threads = check_current_threads();
    await new Promise((resolved) => {
      setTimeout(() => {
        console.log("Thread Polling");
        console.log("Current available threads: %d", available_threads);
        resolved("Poll worker threads");
      }, 2 * 1000);
    });
  }
}

async function crawl(webpages: Array<string>) {
  console.log("crawl");
  try {
    //const database = new WebsiteDatabase().init_database();
    const worker_handler = new WorkerHandler(webpages, MAX_THREADS);
    await worker_handler.crawl_and_index();
    return worker_handler;
  } catch (err) {
    process.exit(1);
  }
}
