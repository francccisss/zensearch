import path from "path";
import { Worker } from "worker_threads";

const THREAD_POOL = 3;

const worker_file = path.join(__dirname, "./Bot/index.ts");

export default class ThreadHandler {
  webpages: Array<string> = [];

  constructor(webpages: Array<string>) {
    this.webpages = webpages;
    this.crawl_and_index();
  }

  private crawl_and_index() {
    const array_buffer = new SharedArrayBuffer(20);
    const shared_buffer = new Uint8Array(array_buffer);

    try {
      for (let i = 0; i < 1; i++) {
        // testing single webpage to index

        const worker = new Worker(worker_file, {
          argv: [i],
          workerData: { shared_buffer },
        });
        console.log(`WorkerID: ${worker.threadId}`);
        worker.on("message", (message) => {
          console.log("Thread changed buffer: ", message);
          worker.postMessage(worker.threadId);
        });
        worker.on("exit", () => {
          console.log(`Worker Exit`);
        });
      }
    } catch (err) {
      const error = err as Error;
      console.log("Log: Something went wrong when creating thread workers.\n");
      console.error(error.message);
    }
  }
}
