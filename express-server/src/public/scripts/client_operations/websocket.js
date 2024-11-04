const ws = new WebSocket("ws://localhost:8080");
import pubsub from "../utils/pubsub.js";

ws.addEventListener("open", (msg) => {
  console.log("Connected to websocket server");
});
ws.addEventListener("message", (event) => {
  const parse_message = JSON.parse(event.data);

  // Response from websocket server
  if (parse_message.message_type === "crawling") {
    pubsub.publish("crawlReceiver", parse_message);
  }
});

function ackMessage() {
  ws.send("ACK");
}

export default { ws, ackMessage };
