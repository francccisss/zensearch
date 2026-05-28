import path from "path";
import rabbitmq from "./rabbitmq/index.js";
import mysql from "mysql2/promise";
import "dotenv/config";
import { exit } from "node:process";
import { readFileSync } from "node:fs";
import { configDotenv } from "dotenv";



configDotenv({ path: path.resolve(import.meta.dirname, "../../.env") })

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
		await execScripts(
			db,
			path.join(import.meta.dirname, "./db_utils/db.init.sql"),
		);

		console.log("tables created");
		console.log("Starting database server");
		const rbqClient = await rabbitmq.EstablishConnection(7);
		await rbqClient.SetDefinitions()
		rabbitmq.SearchEngineHandler(db);
		rabbitmq.DBCheckHandler(db)
		rabbitmq.CrawlerHandler(db);
	} catch (err) {
		const error = err as Error;
		console.error(error);
		exit(1);
	}
})();

async function execScripts(
	db: mysql.Pool | null,
	scriptPath: string,
): Promise<void> {
	console.log(`Executing sql script for ${scriptPath}`);
	if (db === null) {
		console.error("ERROR: database does not exist for %s", scriptPath);
		exit(1);
	}

	const f = readFileSync(scriptPath, "utf-8");
	// TODO: Make sure each table that references a table needs to be triggered right after the
	// referenced table has already been created
	try {
		f.split(";")
			.filter((t) => t.trim())
			.forEach(async (t) => {
				try {
					await db.execute(t);
				} catch (e: any) {
					if (e.message.includes("exists")) {
						console.log("skipping duplicate");
						return;
					}
					console.error(e);
					throw new Error(e);
				}
			});
	} catch (e: any) {
		throw new Error(e);
	}
}
