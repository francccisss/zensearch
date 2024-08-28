import fs from "fs";
import yml from "js-yaml";
import { exit } from "process";

export default function <T>(filePath: string): T {
  try {
    const read = fs.readFileSync(filePath, "utf-8");
    const parseDoc = yml.load(read) as T;
    return parseDoc;
  } catch (err) {
    const error = err as Error;
    console.error("Unable to load file.\n");
    console.error(error.message);
    exit(1);
  }
}
