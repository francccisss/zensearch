import http from "http";
import rabbitmq from "./rabbitmq/index";
import WebsocketService from "./websocket";
import { WebSocketServer } from "ws";
import app from "./express_server";

const PORT = 8080;
(async function start_server() {
  const http_server = http.createServer(app);
  /*
   Connect Rabbitmq
   Creates an indefinite loop to listen/receive
   new messages from the message broker.
  */
  const rbq_client = await rabbitmq.client.connectClient();

  // Connect Websocket for search results retrieved
  const wss: WebSocketServer = new WebSocketServer({ server: http_server });
  const ws_service = new WebsocketService(wss);
  ws_service.handler();

  await rbq_client.init_channel_queues();

  await rbq_client.websocket_channel_listener(
    ws_service.send_crawl_results_to_client.bind(ws_service),
  );
  await rbq_client.search_channel_listener();

  // Start HTTP server
  http_server.listen(PORT, () => {
    console.log("Listening to Port:", PORT);
  });
  /*
   Catching errors propogated by these initializers defined inside
   in the try block
  */
})().catch(console.error);
