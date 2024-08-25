import { EventEmitter } from "stream";
import utils from "./utils";
import ThreadHandler from "./services/Bot/ThreadHandler";
import WebsiteDatabase from "./services/DB";

const event = new EventEmitter();

event.on("crawl", async (webpages: Array<string>) => {
  console.log("crawl");
  console.log(webpages);

  try {
    const thread_handler = new ThreadHandler(
      webpages,
      new WebsiteDatabase().init_database(),
      2,
    );
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
