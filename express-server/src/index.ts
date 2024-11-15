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

  //rbq_client.segmentGenerator();
  await rbq_client.init_channel_queues();

  rbq_client.crawl_channel_listener(
    ws_service.send_crawl_results_to_client.bind(ws_service),
  );
  (async () => {
    let segmentsReceived = 0;
    rbq_client.eventEmitter.on("newSegment", () => {
      segmentsReceived++;
    });

    rbq_client.eventEmitter.on("done", () => {
      console.log("Segments Received: %d", segmentsReceived);
      segmentsReceived = 0;
    });
  })();

  rbq_client.addSegmentsToQueue();
  rbq_client.search_channel_listener();

  // Start HTTP server
  http_server.listen(PORT, () => {
    console.log("Listening to Port:", PORT);
  });
  /*
   Catching errors propogated by these initializers defined inside
   in the try block
  */
})().catch(console.error);
