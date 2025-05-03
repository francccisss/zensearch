import http from "http";
import rabbitmq from "./rabbitmq/index.js";
import WebsocketService from "./websocket/index.js";
import { WebSocketServer } from "ws";
import app from "./express_server/index.js";
import { exit } from "process";

const httpServer = http.createServer(app);
const PORT = 8080;
(async function start_server() {
  console.log("Starting express server");

  httpServer.on("error", (err: any) => {
    if (err.code == "EADDRINUSE") {
      console.error(err.message);
      httpServer.close();
      console.log("exiting process");
      exit(1);
    }
  });
  /*
   Connect Rabbitmq
   Creates an indefinite loop to listen/receive
   new messages from the message broker.
  */
  // const rbqClient = await rabbitmq.client.establishConnection(7);

  // Connect Websocket for search results retrieved
  // const wss: WebSocketServer = new WebSocketServer({ server: httpServer });
  // const wsService = new WebsocketService(wss);
  // wsService.handler();
  //
  // //rbq_client.segmentGenerator();
  // await rbqClient.initChannelQueues();
  //
  // rbqClient.crawlChannelListener(
  //   wsService.sendCrawlResultsToClient.bind(wsService),
  // );
  // (async () => {
  //   let segmentsReceived = 0;
  //   rbqClient.eventEmitter.on("newSegment", () => {
  //     segmentsReceived++;
  //   });
  //
  //   rbqClient.eventEmitter.on("done", () => {
  //     console.log("Segments Received: %d", segmentsReceived);
  //     segmentsReceived = 0;
  //   });
  // })();
  //
  // rbqClient.addSegmentsToQueue();
  // rbqClient.searchChannelListener();

  // Start HTTP server
  httpServer.listen(PORT, () => {
    console.log("Listening to Port:", PORT);
  });
  /*
   Catching errors propogated by these initializers defined inside
   in the try block
  */
})().catch((e) => {
  console.error(e);
  httpServer.close();
  console.log("SHUTDOWN");
  exit(1);
});
