import path from "path";
import { Worker } from "worker_threads";

const worker = new Worker("./src/services/Crawler/index.ts");
