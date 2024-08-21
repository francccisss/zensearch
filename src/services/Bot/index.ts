import { writeFileSync } from "fs";
import { parentPort, workerData } from "worker_threads";
import { Crawler, Scraper } from "./Crawler";
import { exit, title } from "process";
import { Worker } from "cluster";
const current_thread = new Worker();
const scraper = new Scraper();
const crawler = new Crawler(scraper);
const FRAME_SIZE = 1024;
const shared_buffer = new Int32Array(workerData.shared_buffer);
(async function () {
  try {
    //await crawler.start_crawl(process.argv[2]);
    const r_obj = [
      {
        header: {
          title: "This is a title",
          url: "https://doc.python.org/3/",
        },
        webpages: [
          {
            header: {
              title: "Some ttile for this webpage",
              webpage_url: "https://doc.python.org/3/",
            },
            contents: "Some contents in this webpage",
          },

          {
            header: {
              title: "Some ttile for this webpage",
              webpage_url: "https://doc.python.org/3/",
            },
            contents: "Some contents in this webpage",
          },
          {
            header: {
              title: "Some ttile for this webpage",
              webpage_url: "https://doc.python.org/3/",
            },
            webpage_url: "https://doc.python.org/3/",
            contents: "Some contents in this webpage",
          },
        ],
      },
      {
        header: {
          title: "GOLANG",
          url: "https://golang.docs/",
        },
        webpages: [
          {
            header: {
              title: "Some ttile for this webpage",
              webpage_url: "https://doc.python.org/3/",
            },
            webpage_url: "https://golang.docs/",
            contents: "Contents on golang",
          },
          {
            header: {
              title: "Some ttile for this webpage",
              webpage_url: "https://doc.python.org/3/",
            },
            webpage_url: "https://golang.docs/",
            contents: "Contents on golang",
          },
        ],
      },
    ];

    const serialize_obj = JSON.stringify(r_obj[Number(process.argv[2])]);
    const encoder = new TextEncoder();
    const encoded_array = encoder.encode(serialize_obj);
    const encoded_data_buffer = encoded_array.buffer;
    const paddedArray = new Uint8Array(encoded_data_buffer.byteLength * 4); // padding for 32 bit.
    const offset = 4;
    for (let i = 0; i < encoded_array.length; i++) {
      // add the value at every offset
      paddedArray[i * offset] = encoded_array[i];
      // basically leading zeros after the encoded value [<utf-8 value>,0,0,0]
      // if your cpu is using big endianess.. well goodluch xd
    }
    const view = new Int32Array(paddedArray.buffer);
    let current_index = 0;

    while (current_index < view.length) {
      const chunk_size = Math.min(FRAME_SIZE, view.length - current_index);
      for (let i = 0; i < chunk_size; i++) {
        Atomics.store(shared_buffer, i, view[current_index + i]);
      }
      if (current_index <= view.length) {
        Atomics.notify(view, current_index, 1);
      }

      current_index += chunk_size;
    }
    console.log({ current_index, view: view.length });

    parentPort!.postMessage({
      type: "insert",
      shared_buffer,
      data_length: view.length,
    });
  } catch (err) {
    const error = err as Error;
    console.log("LOG: Something went wrong with the current thread.");
    console.error(error.message);
    parentPort!.postMessage({ type: "error" });
  }
})();
