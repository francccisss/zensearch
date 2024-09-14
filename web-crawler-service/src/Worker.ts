import { parentPort, workerData } from "worker_threads";
import { Crawler, Scraper } from "./Crawler";
import { Worker } from "cluster";
const scraper = new Scraper();
const crawler = new Crawler(scraper);
const FRAME_SIZE = 1024;
const shared_buffer = new Int32Array(workerData.shared_buffer);
(async function () {
  try {
    // if crawler crashes, its because of page or browser closing prematurely
    // need to save current dataset.
    const data = await crawler.start_crawl(process.argv[2]);
    if (data === null) throw new Error("Browser closed undexpectedly.");

    const serialize_obj = JSON.stringify(crawler.data);
    const encoder = new TextEncoder();
    const encoded_array = encoder.encode(serialize_obj);
    const encoded_data_buffer = encoded_array.buffer;
    const padded_rray = new Uint8Array(encoded_data_buffer.byteLength * 4); // padding for 32 bit.
    const offset = 4;
    for (let i = 0; i < encoded_array.length; i++) {
      padded_rray[i * offset] = encoded_array[i];
    }
    const view = new Int32Array(padded_rray.buffer);
    let current_index = 0;

    // TODO Fix Atomic operation, for mutliple threads to work with shared buffer

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

    parentPort!.postMessage({
      type: "insert",
      shared_buffer,
      data_length: view.length,
    });
  } catch (err) {
    const error = err as Error;
    console.error("LOG: Current Thread closed unexpectedly.");
    console.error(error.message);
    parentPort!.postMessage({ type: "error", text: error.message });
  }
})();
