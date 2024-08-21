import sqlite3 from "sqlite3";
import path from "path";

class WebsiteDatabase {
  private db: sqlite3.Database | null;
  constructor() {
    this.db = null;
  }
  init_database() {
    if (this.db !== null) {
      return this.db;
    }
    const db_file = "./website_collection.db";
    const sqlite = sqlite3.verbose();
    this.db = new sqlite.Database(
      path.join(__dirname, db_file),
      sqlite.OPEN_READWRITE,
      (err) => {
        if (err) {
          console.error("Unable to connect to website_collection.db");
          process.exit(1);
        }
        console.log("Thread connected to database");
      },
    );
    return this.db;
  }
}

export default WebsiteDatabase;
