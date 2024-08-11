import { EventEmitter } from "stream";
import Crawler from "./services/Crawler";
import utils from "./utils";
import path from "path";

const event = new EventEmitter();

event.on("crawl", (webpages: Array<string>) => {
  console.log("crawl");
  console.log(webpages);
  const worker_import = require("./services/Worker");
});

(function main([_, , ...query_params]: Array<string>) {
  const user_query = query_params[0];
  console.log(user_query);
  const webpages = utils.yaml_loader<{ docs: Array<string> }>(
    "webpage_database.yaml",
  );
  const crawler = new Crawler(webpages.docs);
  event.emit("crawl", crawler.webpages);
})(process.argv);
