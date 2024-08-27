import path from "path";
import { Worker } from "worker_threads";
import { Database, sqlite3 } from "sqlite3";
import { data_t } from "./types/data_t";
import amqp from "amqplib";
import { buffer } from "stream/consumers";

const BUFFER_SIZE = 50000;
const FRAME_SIZE = 1024;
const WORKER_FILE = path.join(__dirname, "./Worker.ts");

type thread_response_t = {
  type: "insert" | "error";
  shared_buffer?: SharedArrayBuffer;
  text?: string;
  data_length: number;
};

export default class ThreadHandler {
  webpages: Array<string> = [];
  db;
  current_threads: number;
  private THREAD_POOL: number;
  successful_thread_count: number;

  constructor(
    webpages: Array<string>,
    database: Database,
    thread_pool: number,
  ) {
    this.webpages = webpages;
    this.db = database;
    this.current_threads = thread_pool; // keep track for polling
    this.THREAD_POOL = thread_pool;
    this.successful_thread_count = 0; // self explanatory
  }

  public async crawl_and_index() {
    const shared_buffer = new SharedArrayBuffer(BUFFER_SIZE);
    try {
      for (let i = 0; i < this.THREAD_POOL; i++) {
        const worker = new Worker(WORKER_FILE, {
          argv: [this.webpages[i]],
          workerData: { shared_buffer },
        });
        this.current_threads--;
        this.event_handlers(worker, shared_buffer);
      }
    } catch (err) {
      this.current_threads++;
      const error = err as Error;
      console.log("Log: Something went wrong when creating thread workers.\n");
      console.error(error.message);
    }
  }

  private async event_handlers(
    worker: Worker,
    shared_buffer: SharedArrayBuffer,
  ) {
    console.log(`WorkerID: ${worker.threadId}`);
    worker.on("message", async (message: thread_response_t) => {
      if (message.type === "error") {
        console.error(
          "Thread %d Threw an error: %s",
          worker.threadId,
          message.text,
        );
        this.current_threads++;
      }

      if (message.type === "insert") {
        if (message.shared_buffer == undefined) return;
        await this.message_decoder(shared_buffer);
        console.log(
          "Thread #%s changed buffer: ",
          worker.threadId,
          shared_buffer.byteLength,
        );
      }
    });
    worker.on("exit", () => {
      console.log(`Worker Exit`);
    });
  }

  private async message_decoder(shared_buffer: SharedArrayBuffer) {
    const view = new Int32Array(shared_buffer);
    let received_chunks = [];
    let current_index = 0;
    // TBH I really dont understand how Atomics work... I need to study more.
    while (current_index < view.length) {
      const chunk_size = Math.min(FRAME_SIZE, view.length - current_index);
      if (view[current_index] !== 0) {
        Atomics.wait(view, current_index, 0);
      }
      received_chunks.push(
        ...view.slice(current_index, current_index + chunk_size),
      );
      current_index += chunk_size;
      for (let i = current_index - chunk_size; i < current_index; i++) {
        Atomics.store(view, i, 0);
      }
    }
    const string_array = new Uint8Array(received_chunks);
    for (let i = 0; i < received_chunks.length; i++) {
      string_array[i] = received_chunks[i];
    }

    const decoder = new TextDecoder();
    const decoded_data = decoder.decode(string_array);
    const last_brace_index = decoded_data.lastIndexOf("}");
    const sliced_object = decoded_data.slice(0, last_brace_index) + "}"; // Buh
    try {
      const deserialize_data = JSON.parse(sliced_object);
      console.log({ received_chunks, sliced_object });
      console.log({ deserialize_data });
      await this.insert_indexed_page(
        deserialize_data,
        string_array.buffer as Buffer,
      );
      this.successful_thread_count++;
    } catch (err) {
      const error = err as Error;
      this.current_threads++;
      console.log("LOG: Decoder was unable to deserialized indexed data.");
      console.error(error.message);
      console.error(error.stack);
    }
  }
  public check_threads() {
    return this.current_threads;
  }

  private async insert_indexed_page(data: data_t, data_buffer: Buffer) {
    console.log(data);
    const connection = await amqp.connect("amqp://localhost");
    const channel = await connection.createChannel();
    const queue = "database_push_queue";

    channel.sendToQueue(queue, data_buffer);
    this.current_threads++;
  }
}
