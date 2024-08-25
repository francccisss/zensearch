import { EventEmitter } from "stream";
import utils from "./utils";
import ThreadHandler from "./ThreadHandler";
import WebsiteDatabase from "./db_interface";
import path from "path";
import amqp, { Channel, Connection } from "amqplib/callback_api";

const event = new EventEmitter();
console.log("Crawl start.");

amqp.connect("amqp://localhost", (err: any, connection: any) => {
  console.log("Connected to MQ");
  if (err) throw err;
  connection.createChannel(function (error: any, channel: Channel) {
    if (error) {
      throw error;
    }
    var queue = "hello";
    channel.assertQueue(queue, {
      durable: false,
    });
    console.log(" [*] Waiting for messages in %s. To exit press CTRL+C", queue);
    channel.consume(
      queue,
      function (msg) {
        if (msg === null) throw new Error("No message");
        console.log(" [x] Received %s", msg!.content.toString());
      },
      {
        noAck: true,
      },
    );
  });
});

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
