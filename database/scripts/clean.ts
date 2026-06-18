import mysql, { type QueryResult } from "mysql2/promise";
import "dotenv/config";
import { configDotenv } from "dotenv";
import path from "node:path";
import { exit } from "node:process";

configDotenv({ path: path.resolve(import.meta.dirname, "../../.env") });

const db = mysql.createPool({
  user: process.env.DB_USER,
  password: process.env.DB_PASS,
  database: process.env.DB_NAME,
  host: process.env.DB_HOST,
});

const [res, _] = await db.execute("show tables;");

for (const table of res as Array<QueryResult>) {
  const tableName = Object.values(table)[0];
  console.log("clean: removing %s", tableName);
  try {
    await db.query(`DELETE FROM ${tableName};`);
  } catch (err: any) {
    console.error(err);
    exit(1);
  }
}
console.log("Current tables:");
console.log(res);

console.log("Done");
exit(0);
