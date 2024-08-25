import { parentPort, workerData } from "worker_threads";
import { Crawler, Scraper } from "./Crawler";
import { Worker } from "cluster";
const scraper = new Scraper();
const crawler = new Crawler(scraper);
const FRAME_SIZE = 1024;
const shared_buffer = new Int32Array(workerData.shared_buffer);
(async function () {
  try {
    await crawler.start_crawl(process.argv[2]);
    const serialize_obj = JSON.stringify(crawler.data);
    const encoder = new TextEncoder();
    const encoded_array = encoder.encode(serialize_obj);
    const encoded_data_buffer = encoded_array.buffer;
    const padded_rray = new Uint8Array(encoded_data_buffer.byteLength * 4); // padding for 32 bit.
    const offset = 4;
    for (let i = 0; i < encoded_array.length; i++) {
      // add the value at every offset
      padded_rray[i * offset] = encoded_array[i];
      // basically leading zeros after the encoded value [<utf-8 value>,0,0,0]
      // if your cpu is using big endianess.. well goodluch xd
    }
    const view = new Int32Array(padded_rray.buffer);
    let current_index = 0;

    console.log("AFTER ENCODING");
    console.log({
      encoded_data: encoded_array,
      last_index: encoded_array[2452],
      length: encoded_array.length,
    });

    while (current_index < view.length) {
      const chunk_size = Math.min(FRAME_SIZE, view.length - current_index);
      for (let i = 0; i < chunk_size; i++) {
        Atomics.store(
          shared_buffer,
          current_index + i,
          view[current_index + i],
        );
      }
      if (current_index <= view.length) {
        Atomics.notify(view, current_index, 1);
      }
      current_index += chunk_size;
    }

    console.log("AFTER TRANFERRING ENCODED DATA TO SHARED BUFFER");
    console.log({
      shared_buffer: shared_buffer.slice(0, 100),
      last_index: shared_buffer[2452],
      length: shared_buffer.length,
    });

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
