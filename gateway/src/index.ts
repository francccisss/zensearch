import http from "http";
import rabbitmq from "./rabbitmq/index";
import WebsocketService from "./websocket";
import { WebSocketServer } from "ws";
import app from "./express_server";

const PORT = 8080;
(async function start_server() {
  const http_server = http.createServer(app);
  // Connect Rabbitmq
  const message_broker = await rabbitmq.connect();
  // Connect Websocket for search results retreival
  const wss: WebSocketServer = new WebSocketServer({ server: http_server });
  const ws_service = new WebsocketService(wss);
  ws_service.handler();
  await rabbitmq.init_search_channel_queues();
  await rabbitmq.search_listener((data) => {
    ws_service.send_results_to_client(data);
  });

  // Start HTTP server
  http_server.listen(PORT, () => {
    console.log("Listening to Port:", PORT);
  });
})().catch(console.error);
