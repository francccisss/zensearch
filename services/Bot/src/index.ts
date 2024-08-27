import { EventEmitter } from "stream";
import utils from "./utils";
import ThreadHandler from "./ThreadHandler";
import WebsiteDatabase from "./db_interface";
import path from "path";
import amqp, { Channel, Connection } from "amqplib";

const event = new EventEmitter();
const MAX_THREADS = 1;
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
      const main_thread = await crawl(to_json.docs);
      let available_threads = main_thread.current_threads;
      await thread_polling(
        available_threads,
        main_thread.check_threads.bind(main_thread),
      );
      console.log(available_threads);
      channel.sendToQueue(
        msg.properties.replyTo,
        Buffer.from(`Crawled: ${to_json.docs}`),
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
    const database = new WebsiteDatabase().init_database();
    const thread_handler = new ThreadHandler(webpages, database, MAX_THREADS);
    await thread_handler.crawl_and_index();
    return thread_handler;
  } catch (err) {
    process.exit(1);
  }
}
