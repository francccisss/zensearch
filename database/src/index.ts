import path from "path";
import rabbitmq from "./rabbitmq/index.js";
import mysql from "mysql2/promise";
import "dotenv/config";
import { exit } from "node:process";
import { configDotenv } from "dotenv";

configDotenv({ path: path.resolve(import.meta.dirname, "../../.env") });

// TODO: UNABLE TO RESOLVE .ENV FILE
const poolOption: mysql.PoolOptions = {
  user: process.env.DB_USER,
  password: process.env.DB_PASS,
  database: process.env.DB_NAME,
  host: process.env.DB_HOST,
  multipleStatements: false,
};

await (async function init(): Promise<void> {
  try {
    const db = mysql.createPool(poolOption);

    console.log("Starting database server");
    const rbqClient = await rabbitmq.EstablishConnection(7);
    await rbqClient.SetDefinitions();
    rabbitmq.SearchEngineHandler(db);
    rabbitmq.DBCheckHandler(db);
    rabbitmq.CrawlerHandler(db);
  } catch (err) {
    const error = err as Error;
    console.error(error);
    exit(1);
  }
})();
