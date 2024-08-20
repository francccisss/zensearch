import { writeFileSync } from "fs";
import { parentPort, workerData } from "worker_threads";
import { Crawler, Scraper } from "./Crawler";
import { exit } from "process";
import { Worker } from "cluster";
import { arrayBuffer } from "stream/consumers";
import { encode } from "punycode";
const current_thread = new Worker();
const scraper = new Scraper();
const crawler = new Crawler(scraper);
const FRAME_SIZE = 1024;
const shared_buffer = new Int32Array(workerData.shared_buffer);
(async function () {
  try {
    //await crawler.start_crawl(process.argv[2]);
    const r_obj = {
      header: {
        title: "This is a title",
      },
      url: "https://doc.python.org/3/",
      webpages: [
        {
          webpage_url: "https://doc.python.org/3/",
          contents: "Some contents in this webpage",
        },

        {
          webpage_url: "https://doc.python.org/3/",
          contents: "Some contents in this webpage",
        },
        {
          webpage_url: "https://doc.python.org/3/",
          contents: "Some contents in this webpage",
        },
      ],
    };

    const serialize_obj = JSON.stringify(r_obj);
    const encoder = new TextEncoder();
    const byte_array = encoder.encode(serialize_obj);
    const encoded_data_buffer = byte_array.buffer;
    const paddingLength = (4 - (byte_array.length % 4)) % 4;
    const paddedArray = new Uint8Array(byte_array.length + paddingLength);
    paddedArray.set(byte_array, 0);

    const view = new Int32Array(paddedArray.buffer);
    let current_index = 0;

    while (current_index < view.length) {
      const chunk_size = Math.min(FRAME_SIZE, view.length - current_index);
      for (let i = 0; i < chunk_size; i++) {
        Atomics.store(shared_buffer, i, view[current_index + i]);
      }
      current_index += chunk_size;
      if (current_index === view.length) {
        Atomics.notify(shared_buffer, current_index, 1);
      }
    }
    console.log({ current_index, view: view.length });

    parentPort!.postMessage({ type: "insert", shared_buffer });
    parentPort!.on("message", (value) => {
      console.log(value);
    });
  } catch (err) {
    const error = err as Error;
    console.log("LOG: Something went wrong with the current thread.");
    console.error(error.message);
    parentPort!.postMessage({ type: "error" });
  }
})();
