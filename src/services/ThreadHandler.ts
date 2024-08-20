import path from "path";
import { Worker } from "worker_threads";
import { data_t } from "../types/data_t";
import WebsiteDatabase from "./DB";
import { arrayBuffer } from "stream/consumers";
const THREAD_POOL = 3;
const BUFFER_SIZE = 2048;
const worker_file = path.join(__dirname, "./Bot/index.ts");

type thread_response_t = {
  type: "insert" | "error";
  shared_buffer?: ArrayBufferLike;
};

export default class ThreadHandler {
  webpages: Array<string> = [];
  db;

  constructor(webpages: Array<string>) {
    this.webpages = webpages;
    this.db = new WebsiteDatabase().init_database();
    this.crawl_and_index();
  }

  private event_handlers(worker: Worker, shared_buffer: ArrayBufferLike) {
    console.log(`WorkerID: ${worker.threadId}`);
    worker.on("message", (message: thread_response_t) => {
      if (message.type === "error") {
        console.log("Thread s% Threw an error: ", worker.threadId);
      }

      if (message.type === "insert") {
        if (message.shared_buffer == undefined) return;
        this.message_decoder(message.shared_buffer);
        console.log(
          "Thread %s changed buffer: ",
          worker.threadId,
          message.shared_buffer,
        );
      }
    });
    worker.on("exit", () => {
      console.log(`Worker Exit`);
    });
  }

  private message_decoder(thread_buffer: ArrayBufferLike) {
    const decoder = new TextDecoder();
    const decoded_data = decoder.decode(thread_buffer);
    const last_brace_index = decoded_data.lastIndexOf("}");
    const sliced_object = decoded_data.slice(0, last_brace_index + 1);
    const deserialize_data = JSON.parse(sliced_object);
    console.log(deserialize_data);
    this.insert_indexed_page(deserialize_data);
  }

  private crawl_and_index() {
    const array_buffer = new SharedArrayBuffer(BUFFER_SIZE);
    const shared_buffer = new Uint8Array(array_buffer);
    try {
      for (let i = 0; i < 1; i++) {
        const worker = new Worker(worker_file, {
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
