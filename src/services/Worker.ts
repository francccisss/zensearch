import path from "path";
const { Worker } = require("worker_threads");

const worker_file = path.join(__dirname, "./Crawler/index.ts");
const worker = new Worker(worker_file);
