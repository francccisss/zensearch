import { writeFileSync } from "fs";
import { parentPort, workerData } from "worker_threads";
import sqlite3 from "sqlite3";
import path from "path";
import { Crawler, Scraper } from "./Crawler";

const scraper = new Scraper();
const crawler = new Crawler(scraper);
const FRAME_SIZE = 2;
const shared_buffer = new Uint8Array(workerData.shared_buffer);

const sqlite = sqlite3.verbose();
const db = new sqlite.Database(
  path.join(__dirname, "../DB/indexed_webpages.db"),
  (err) => {
    if (err) {
      console.error("Unable to connect to indexed_webpages.db");
      return;
    }
    console.log("Connected to database from a different process");
  },
);

let current_index = 0;
for (let i = current_index; i < 4; i++) {
  for (let j = 0; j < FRAME_SIZE; j++) {
    shared_buffer[i] = i;
  }
  current_index++;
}

console.log(shared_buffer);
parentPort!.postMessage(shared_buffer);
parentPort!.on("message", (value) => {
  console.log(value);
});
