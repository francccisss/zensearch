const ws = new WebSocket("ws://192.168.1.20:8080");
import pubsub from "../utils/pubsub.js";

ws.addEventListener("open", (msg) => {
  console.log("Connected to websocket server");
});
ws.addEventListener("message", (event) => {
  const parse_message = JSON.parse(event.data);

  // Server side's `CRAWL_CB_QUEUE` consumer will push crawled
  // websites to this message_type of type `crawling`.

  // for each crawled websites on the crawl list that was sent to the
  // server, we need to store each one in a global variable, and if
  // the length of the messages received from the server is equal
  // to the length of the crawl_list length, then transition to onSuccessCrawl

  // TODO Error: Receiving message just right after sending the unindexed list
  // should only receive after the crawl is done.
  if (parse_message.message_type === "crawling") {
    console.log("Message received from crawler");
    pubsub.publish("crawlReceiver", parse_message);
  }
});

export default ws;
