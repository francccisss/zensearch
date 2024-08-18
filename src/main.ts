import { EventEmitter } from "stream";
import utils from "./utils";
import path from "path";
import ThreadHandler from "./services/ThreadHandler";

const event = new EventEmitter();

event.on("crawl", async (webpages: Array<string>) => {
  console.log("crawl");
  console.log(webpages);
  try {
    const worker = new ThreadHandler(webpages);
  } catch (err) {
    process.exit(1);
  }
});

(function main([_, , ...query_params]: Array<string>) {
  const user_query = query_params[0];
  const webpages = utils.yaml_loader<{ docs: Array<string> }>(
    "webpage_database.yaml",
  );
  event.emit("crawl", webpages.docs);
})(process.argv);
