import http from "http";
import rabbitmq from "./rabbitmq/index";
import WebsocketService from "./websocket";
import { WebSocketServer } from "ws";
import app from "./express_server";

const PORT = 8080;
(async function start_server() {
  const httpServer = http.createServer(app);
  /*
   Connect Rabbitmq
   Creates an indefinite loop to listen/receive
   new messages from the message broker.
  */
  const rbqClient = await rabbitmq.client.establishConnection(7);

  // Connect Websocket for search results retrieved
  const wss: WebSocketServer = new WebSocketServer({ server: httpServer });
  const wsService = new WebsocketService(wss);
  wsService.handler();

  //rbq_client.segmentGenerator();
  await rbqClient.initChannelQueues();

  rbqClient.crawlChannelListener(
    wsService.sendCrawlResultsToClient.bind(wsService),
  );
  (async () => {
    let segmentsReceived = 0;
    rbqClient.eventEmitter.on("newSegment", () => {
      segmentsReceived++;
    });

    rbqClient.eventEmitter.on("done", () => {
      console.log("Segments Received: %d", segmentsReceived);
      segmentsReceived = 0;
    });
  })();

  rbqClient.addSegmentsToQueue();
  rbqClient.searchChannelListener();

  // Start HTTP server
  httpServer.listen(PORT, () => {
    console.log("Listening to Port:", PORT);
  });
  /*
   Catching errors propogated by these initializers defined inside
   in the try block
  */
})().catch(console.error);
