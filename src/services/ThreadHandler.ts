import path from "path";
import { Worker } from "worker_threads";
import { data_t } from "../types/data_t";
import WebsiteDatabase from "./DB";
import { arrayBuffer } from "stream/consumers";
const db = new WebsiteDatabase().init_database();
const THREAD_POOL = 3;
const worker_file = path.join(__dirname, "./Bot/index.ts");

type thread_response_t = {
  type: "insert" | "error";
  shared_buffer?: ArrayBufferLike;
};

export default class ThreadHandler {
  webpages: Array<string> = [];

  constructor(webpages: Array<string>) {
    this.webpages = webpages;
    this.crawl_and_index();
  }
  private event_handlers(worker: Worker) {
    console.log(`WorkerID: ${worker.threadId}`);
    worker.on("message", (message: thread_response_t) => {
      if (message.type === "error") {
        console.log("Thread #%s Threw an error: ");
      }
      if (message.type === "insert") {
        console.log(
          "Thread #%s changed buffer: ",
          worker.threadId,
          message.shared_buffer,
        );
      }
    });
    worker.on("exit", () => {
      console.log(`Worker Exit`);
    });
  }
  private message_decoder(message: ArrayBufferLike) {
    const decoder = new TextDecoder();
    const decoded_data = decoder.decode(message);
  }

  private crawl_and_index() {
    const array_buffer = new SharedArrayBuffer(20);
    const shared_buffer = new Uint8Array(array_buffer);
    try {
      for (let i = 0; i < 2; i++) {
        const worker = new Worker(worker_file, {
          argv: [i],
          workerData: { shared_buffer },
        });
        this.event_handlers(worker);
      }
    } catch (err) {
      const error = err as Error;
      console.log("Log: Something went wrong when creating thread workers.\n");
      console.error(error.message);
    }
  }
  private insert_indexed_page(data: data_t) {}
}
