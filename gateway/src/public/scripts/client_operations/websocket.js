const ws = new WebSocket("ws://192.168.1.20:8080");
import pubsub from "../utils/pubsub.js";

ws.addEventListener("open", (msg) => {
  console.log("Connected to websocket server");
});
ws.addEventListener("message", (event) => {
  /*
   Just a work around for sending ping pong shenanigans
   because client side doesn't support sending an on pong event
   so the websocket server has to send a "pong" back to client,
   when in ackshuality the server should be the one to listen
   to pong events from the client side, so when server sends data (ping)
   the client should send a pong with an ACK message, but instead
   the client CAN only send a ping, so when the server receives the "ping"
   the server then emits out a "pong" event with an "ACK" message, which
   the rabbitmq handler will read that the client side has indeed received
   the message.
  */
  //if (event.data.toString() === "ACK") {
  //  console.log("Ack");
  //  return;
  //}
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

function ackMessage() {
  console.log("ACK SENT FROM CLIENT");
  ws.send("ACK");
}

export default { ws, ackMessage };
