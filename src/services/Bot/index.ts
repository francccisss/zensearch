import { writeFileSync } from "fs";
import { parentPort, workerData } from "worker_threads";
import sqlite3 from "sqlite3";
import path from "path";
import { Crawler, Scraper } from "./Crawler";
import { exit } from "process";

const db_file_endpoint = "../DB/indexed_webpages.db";
const scraper = new Scraper();
const crawler = new Crawler(scraper);
const FRAME_SIZE = 2;
const shared_buffer = new Uint8Array(workerData.shared_buffer);

const sqlite = sqlite3.verbose();
const db = new sqlite.Database(
  path.join(__dirname, db_file_endpoint),
  sqlite.OPEN_READWRITE,
  (err) => {
    if (err) {
      console.error("Unable to connect to indexed_webpages.db");
      exit(1);
    }
    console.log("Thread connected to database");
  },
);

let current_index = 0;
for (let i = current_index; i < 4; i++) {
  for (let j = 0; j < FRAME_SIZE; j++) {
    shared_buffer[i] = i;
  }
  current_index++;
}

parentPort!.postMessage(shared_buffer);
parentPort!.on("message", (value) => {
  console.log(value);
});
