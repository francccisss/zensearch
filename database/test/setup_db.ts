import Database from "better-sqlite3";
import { readFile } from "fs";
import { exit } from "process";

const tables = [
  "known_sites",
  "indexed_sites",
  "webpages",
  "visited_nodes", // Dont move this before node
  "nodes",
  "queues",
];

export function initDatabase(src: string): Database.Database {
  const db = new Database(src);
  return db;
}

export function execScripts(db: Database.Database | null, scriptPath: string) {
  console.log("Execute sqlite script");
  console.log(scriptPath);
  if (db === null) {
    console.error("ERROR: database does not exist for %s", scriptPath);
    exit(1);
  }
  readFile(scriptPath, "utf-8", (_, data) => {
    const stmts = data
      .split(";")
      .map((stmt) => stmt.trim())
      .filter((stmt) => stmt);
    stmts.forEach((stmt) => {
      const firstLine = stmt.split("\n")[0];
      for (let i = 0; i < tables.length; i++) {
        const tableName = tables[i];
        if (firstLine.includes(tableName)) {
          console.log(tableName);
          const checkTableStmt = db.prepare(
            `SELECT name FROM sqlite_master WHERE type='table' AND name=? ;`,
          );
          const result = checkTableStmt.get([tableName]);
          if (result == undefined) {
            console.log("Notif: Creating table for %s", tableName);
            db.exec(stmt);
            break;
          }
        }
      }
    });
  });
}
