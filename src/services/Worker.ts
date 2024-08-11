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
    for (let i = 0; i < this.webpages.length; i++) {
      const worker = new Worker(worker_file, { argv: [this.webpages[i]] });
    }
  }
}
