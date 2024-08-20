import path from "path";
import { Worker } from "worker_threads";
import { data_t } from "../types/data_t";
import WebsiteDatabase from "./DB";
import { arrayBuffer } from "stream/consumers";

const THREAD_POOL = 3;
const BUFFER_SIZE = 5048;
const FRAME_SIZE = 1024;
const WORKER_FILE = path.join(__dirname, "./Bot/index.ts");

type thread_response_t = {
  type: "insert" | "error";
  shared_buffer?: SharedArrayBuffer;
};

export default class ThreadHandler {
  webpages: Array<string> = [];
  db;

  constructor(webpages: Array<string>) {
    this.webpages = webpages;
    this.db = new WebsiteDatabase().init_database();
    this.crawl_and_index();
  }

  private event_handlers(worker: Worker, shared_buffer: SharedArrayBuffer) {
    console.log(`WorkerID: ${worker.threadId}`);
    worker.on("message", (message: thread_response_t) => {
      if (message.type === "error") {
        console.log("Thread s% Threw an error: ", worker.threadId);
      }

      if (message.type === "insert") {
        if (message.shared_buffer == undefined) return;
        this.message_decoder(shared_buffer);
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

  private message_decoder(shared_buffer: SharedArrayBuffer) {
    // size in bytes is 5048 each occupied 32-bit is 4 bytes
    // in little endian byte ordering
    const view = new Int32Array(shared_buffer);
    let received_chunks = [];
    let current_index = 0;
    while (current_index < view.length) {
      received_chunks.push(
        ...view.slice(current_index, current_index + FRAME_SIZE),
      );
      current_index += FRAME_SIZE;
      Atomics.wait(view, 0, 0);
    }
    const string_array = new Uint8Array(received_chunks);
    for (let i = 0; i < received_chunks.length; i++) {
      string_array[i] = received_chunks[i];
    }

    console.log({ received_chunks, string_array, view });
    const decoder = new TextDecoder();
    const decoded_data = decoder.decode(string_array);
    const last_brace_index = decoded_data.lastIndexOf("}");
    const sliced_object = decoded_data.slice(0, last_brace_index + 1);
    console.log({ received_chunks, string_array, sliced_object });
    const deserialize_data = JSON.parse(sliced_object);

    console.log({ deserialize_data });
    this.insert_indexed_page(deserialize_data);
  }

  private crawl_and_index() {
    const shared_buffer = new SharedArrayBuffer(BUFFER_SIZE);
    try {
      for (let i = 0; i < 1; i++) {
        const worker = new Worker(WORKER_FILE, {
          argv: [i],
          workerData: { shared_buffer },
        });
        this.event_handlers(worker, shared_buffer);
      }
    } catch (err) {
      const error = err as Error;
      console.log("Log: Something went wrong when creating thread workers.\n");
      console.error(error.message);
    }
  }
  private insert_indexed_page(data: data_t) {}
}
