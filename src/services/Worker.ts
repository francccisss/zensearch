import path from "path";
import { Worker } from "worker_threads";

const THREAD_POOL = 3;

const worker_file = path.join(__dirname, "./Crawler/index.ts");

export default class WorkerThread {
  webpages: Array<string> = [];

  constructor(webpages: Array<string>) {
    this.webpages = webpages;
    this.init();
  }

  private init() {
    try {
      for (let i = 0; i < this.webpages.length - 3; i++) {
        // testing single webpage to index
        const worker = new Worker(worker_file, { argv: [this.webpages[i]] });
        console.log(`WorkerID: ${worker.threadId}`);
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
