import http from "http";
import rabbitmq from "./rabbitmq/index.js";
import WebsocketService from "./websocket/index.js";
import { WebSocketServer } from "ws";
import app from "./app/index.js";
import { exit } from "process";
import EventEmitter from "events";

const deferEvent = new EventEmitter({});
deferEvent.addListener("dc", async () => {
  console.log("Server shutting down, removing ephemeral queues");

  const [v, c, d] = await Promise.allSettled([
    rabbitmq.eventsChannel!.deleteQueue(
      rabbitmq.definitions.queues.es_se_query_cbq,
      { ifUnused: false, ifEmpty: false },
    ),
    rabbitmq.eventsChannel!.deleteQueue(
      rabbitmq.definitions.queues.es_db_check_cbq,

      { ifUnused: false, ifEmpty: false },
    ),
    rabbitmq.eventsChannel!.deleteQueue(
      rabbitmq.definitions.queues.es_cr_request_cbq,
      { ifUnused: false, ifEmpty: false },
    ),
  ]);
  console.log(
    `Status: es_se_qeury_cbq: ${v.status} - es_db_check_cbq: ${c.status} - es_cr_request_cbq: ${d.status}`,
  );
  console.log("Done");
});

process.on("SIGINT", () => {
  const successDefer = deferEvent.emit("dc");
  if (successDefer != true) {
    console.log("Unable to call dc event, ephemeral queues are still up ");
  }
});

const httpServer = http.createServer(app);
const PORT = 8080;
await (async function start_server() {
  console.log(import.meta.dirname);

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
  const rbqConn = await rabbitmq.client.EstablishConnection(7);

  // Connect Websocket for search results retrieved
  const wss: WebSocketServer = new WebSocketServer({ server: httpServer });
  const wsService = new WebsocketService(wss);
  wsService.handler();

  await rbqConn.SetDefinitions();

  rbqConn.CrawlChannelHandler(
    wsService.sendCrawlResultsToClient.bind(wsService),
  );
  (async () => {
    let segmentsReceived = 0;
    rbqConn.eventEmitter.on("newSegment", () => {
      segmentsReceived++;
    });

    rbqConn.eventEmitter.on("done", () => {
      console.log("Segments Received: %d", segmentsReceived);
      segmentsReceived = 0;
    });
  })();

  rbqConn.AddSegmentsToQueue();
  rbqConn.SearchChannelHandler();

  // Start HTTP server
  httpServer.listen(PORT, () => {
    console.log("Listening to Port:", PORT);
  });
  /*
	 Catching errors propogated by these initializers defined inside
	 in the try block
	*/
})().catch((e: any) => {
  console.error(e);
  httpServer.close();
  console.log("SHUTDOWN");
  const successDefer = deferEvent.emit("dc");
  if (successDefer != true) {
    console.log("Unable to call dc event, ephemeral queues are still up ");
  }
  exit(1);
});
