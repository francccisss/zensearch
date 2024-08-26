import { EventEmitter } from "stream";
import utils from "./utils";
import ThreadHandler from "./ThreadHandler";
import WebsiteDatabase from "./db_interface";
import path from "path";
import amqp, { Channel, Connection } from "amqplib";

const event = new EventEmitter();
console.log("Crawl start.");

(async function () {
  const connection = await amqp.connect("amqp://localhost");
  const channel = await connection.createChannel();
  var queue = "crawl_rpc_queue";
  await channel.assertQueue(queue, {
    durable: false,
  });
  console.log(" [*] Waiting for messages in %s. To exit press CTRL+C", queue);
  const decoder = new TextDecoder();

  await channel.consume(
    queue,
    function (msg) {
      if (msg === null) throw new Error("No message");
      const decoded_array_buffer = decoder.decode(msg.content);
      const to_json = JSON.parse(decoded_array_buffer);
      console.log(to_json);
      setTimeout(() => {
        channel.sendToQueue(msg.properties.replyTo, Buffer.from("CRAAAWLEED"), {
          correlationId: msg.properties.correlationId,
        });
      }, 5000);
    },
    {
      noAck: true,
    },
  );
})();
//event.on("crawl", async (webpages: Array<string>) => {
//  console.log("crawl");
//  console.log(webpages);
//
//  try {
//    const thread_handler = new ThreadHandler(
//      webpages,
//      new WebsiteDatabase().init_database(),
//      2,
//    );
//  } catch (err) {
//    process.exit(1);
//  }
//});
//
//(function main([_, , ...query_params]: Array<string>) {
//  const user_query = query_params[0];
//  const webpages = utils.yaml_loader<{ docs: Array<string> }>(
//    path.join(__dirname, "webpage_database.yaml"),
//  );
//  event.emit("crawl", webpages.docs);
//})(process.argv);
