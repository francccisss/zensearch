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
    const encoded_array = encoder.encode(serialize_obj);
    const encoded_data_buffer = encoded_array.buffer;
    const paddedArray = new Uint8Array(encoded_data_buffer.byteLength * 4); // padding for 32 bit.
    const offset = 4;
    for (let i = 0; i < encoded_array.length; i++) {
      // add the value at every offset
      paddedArray[i * offset] = encoded_array[i];
    }
    //7b 22 68 65 will be a single 32bit word
    //7b 00 00 00 <- needs to be this one
    //need offset of 4
    //Imm(bp,index,offset)
    //index * offset = encoded[i]

    console.log({ encoded_data_buffer, paddedArray });
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
