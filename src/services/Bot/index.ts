import { writeFileSync } from "fs";
import { parentPort, workerData } from "worker_threads";
import { Crawler, Scraper } from "./Crawler";
import { exit } from "process";
import { Worker } from "cluster";
const current_thread = new Worker();
const scraper = new Scraper();
const crawler = new Crawler(scraper);
const FRAME_SIZE = 2;
const shared_buffer = new Uint8Array(workerData.shared_buffer);
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
    const array_buffer = byte_array.buffer;

    const view = new Uint8Array(array_buffer);

    const decoder = new TextDecoder();
    const decoded_data = decoder.decode(view);

    let current_index = 0;
    for (let i = current_index; i < 4; i++) {
      for (let j = 0; j < FRAME_SIZE; j++) {
        shared_buffer[i] = i;
      }
      current_index++;
    }

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
